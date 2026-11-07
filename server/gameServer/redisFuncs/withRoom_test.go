package redisFuncs

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newWithRoomRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

func seedRoom(t *testing.T, client *redis.Client, roomID string) {
	t.Helper()
	room := rooms.Room{ID: roomID}
	data, err := json.Marshal(room)
	if err != nil {
		t.Fatalf("marshal seed: %v", err)
	}
	if err := client.Set(context.Background(), "room:"+roomID, data, 0).Err(); err != nil {
		t.Fatalf("seed room: %v", err)
	}
}

func TestWithRoomDoisPods(t *testing.T) {
	_, client := newWithRoomRedis(t)
	ctx := context.Background()
	const roomID = "sala-dois-pods"
	seedRoom(t, client, roomID)
	t.Setenv("ROOM_LOCK_TTL_MS", "3000")

	// Pod-a adquire o lock e grava um estado; avisa quando já está dentro do fn.
	config.InstanceID = "pod-a"
	podAStarted := make(chan struct{})
	podADone := make(chan error, 1)

	go func() {
		err := WithRoom(ctx, client, roomID, CausePlay, func(room *rooms.Room) bool {
			close(podAStarted)
			room.Turn = 42
			return true
		})
		podADone <- err
	}()

	<-podAStarted

	// Troca para o segundo pod: ele espera o primeiro terminar e vê o estado gravado.
	config.InstanceID = "pod-b"
	var sawTurn int
	err := WithRoom(ctx, client, roomID, CausePlay, func(room *rooms.Room) bool {
		sawTurn = room.Turn
		return false
	})
	if err != nil {
		t.Fatalf("pod-b WithRoom: %v", err)
	}
	if err := <-podADone; err != nil {
		t.Fatalf("pod-a WithRoom: %v", err)
	}
	if sawTurn != 42 {
		t.Fatalf("pod-b viu Turn=%d, quer 42", sawTurn)
	}
}

func TestWithRoomNaoGravaQuandoFnDevolveFalse(t *testing.T) {
	_, client := newWithRoomRedis(t)
	ctx := context.Background()
	const roomID = "sala-nao-aplicado"
	seedRoom(t, client, roomID)

	err := WithRoom(ctx, client, roomID, CauseTick, func(room *rooms.Room) bool {
		room.Turn = 999 // muta, mas o tick "não rodou"
		return false
	})
	if err != nil {
		t.Fatalf("WithRoom: %v", err)
	}

	fresh, err := LoadRoom(ctx, client, roomID)
	if err != nil {
		t.Fatalf("LoadRoom: %v", err)
	}
	if fresh.Turn != 0 {
		t.Fatalf("fn devolveu false mas a mutação foi gravada: Turn=%d", fresh.Turn)
	}
}

func TestWithRoomPerdaDoLockCancelaSave(t *testing.T) {
	_, client := newWithRoomRedis(t)
	ctx := context.Background()
	const roomID = "sala-perda-save"
	seedRoom(t, client, roomID)
	t.Setenv("ROOM_LOCK_TTL_MS", "300")

	config.InstanceID = "pod-a"
	saveResult := make(chan error, 1)

	go func() {
		err := WithRoom(ctx, client, roomID, CausePlay, func(room *rooms.Room) bool {
			// Simula a perda do lock apagando a chave.
			if err := client.Del(ctx, "lock:room:"+roomID).Err(); err != nil {
				saveResult <- err
				return false
			}
			// Espera a renovação detectar e cancelar o ctx: o SaveRoom do WithRoom deve falhar.
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				if room.Lock().Err() != nil {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			return true
		})
		saveResult <- err
	}()

	err := <-saveResult
	if err == nil {
		t.Fatal("WithRoom deveria falhar ao salvar após a perda do lock")
	}
}

// commandCounter conta comandos/round-trips vistos pelo client do Redis.
type commandCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func newCommandCounter() *commandCounter {
	return &commandCounter{counts: make(map[string]int)}
}

func (c *commandCounter) count(cmds ...redis.Cmder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cmd := range cmds {
		c.counts[cmd.Name()]++
	}
}

func (c *commandCounter) get(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[name]
}

func (c *commandCounter) DialHook(next redis.DialHook) redis.DialHook { return next }

func (c *commandCounter) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		c.count(cmd)
		return next(ctx, cmd)
	}
}

func (c *commandCounter) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		c.count(cmds...)
		return next(ctx, cmds)
	}
}

func TestFillConnectedUsaMGetEPreservaSemantica(t *testing.T) {
	_, client := newWithRoomRedis(t)
	ctx := context.Background()
	const roomID = "sala-connected"

	// p1: nesta sala, instância viva (conectado).
	// p2: em outra sala (não conectado).
	// p3: valor malformado (registro removido).
	// p4: instância morta (registro removido).
	// p5: sem registro.
	for key, value := range map[string]string{
		"instance:pod-a:alive": "1",
		"player:p1:room":       roomID + ":pod-a",
		"player:p2:room":       "outra-sala:pod-a",
		"player:p3:room":       "sem-separador",
		"player:p4:room":       roomID + ":pod-morto",
	} {
		if err := client.Set(ctx, key, value, 0).Err(); err != nil {
			t.Fatalf("seed %s: %v", key, err)
		}
	}

	room := &rooms.Room{
		ID: roomID,
		Players: map[string]*players.Player{
			"p1": {Id: "p1"},
			"p2": {Id: "p2"},
			"p3": {Id: "p3"},
			"p4": {Id: "p4"},
			"p5": {Id: "p5"},
		},
	}

	counter := newCommandCounter()
	client.AddHook(counter)

	if err := fillConnected(ctx, client, room); err != nil {
		t.Fatalf("fillConnected: %v", err)
	}

	if len(room.Connected) != 1 || room.Connected[0] != "p1" {
		t.Fatalf("connected = %v, quer [p1]", room.Connected)
	}

	// Mesma semântica do caminho por jogador: registros malformados/órfãos saem.
	for _, playerId := range []string{"p3", "p4"} {
		exists, err := client.Exists(ctx, "player:"+playerId+":room").Result()
		if err != nil {
			t.Fatalf("exists %s: %v", playerId, err)
		}
		if exists != 0 {
			t.Fatalf("registro de %s deveria ter sido removido", playerId)
		}
	}
	if exists, _ := client.Exists(ctx, "player:p2:room").Result(); exists != 1 {
		t.Fatal("registro de p2 (outra sala) não deveria ser removido")
	}

	// Sem GET por jogador: só MGETs (registros e instâncias).
	if got := counter.get("get"); got != 0 {
		t.Fatalf("fillConnected fez %d GET(s), quer 0", got)
	}
	if got := counter.get("mget"); got > 2 {
		t.Fatalf("fillConnected fez %d MGET(s), quer no máximo 2", got)
	}
}
