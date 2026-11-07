package services

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newReplayRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

func testSeed() rooms.Seed {
	var s rooms.Seed
	for i := range s {
		s[i] = byte(i + 1)
	}
	return s
}

func newReplayRoom(id, matchID string, started bool) *rooms.Room {
	room := &rooms.Room{
		ID:         id,
		Rules:      structs.ClassicRules,
		MatchID:    matchID,
		Seed:       testSeed(),
		Rng:        rooms.NewRngState(testSeed()),
		Turn:       1,
		MaxPlayers: 10,
		Players: map[string]*players.Player{
			"p1": {Id: "p1", Name: "Ana", MoveTime: 2, Alive: true},
			"p2": {Id: "p2", Name: "Bia", MoveTime: 1, Alive: true},
		},
	}
	if started {
		room.StartTime = time.Now().UTC()
	}
	return room
}

func seedReplayRoom(t *testing.T, client *redis.Client, room *rooms.Room) {
	t.Helper()
	if err := redisFuncs.SaveRoom(context.Background(), client, room); err != nil {
		t.Fatalf("seed room: %v", err)
	}
}

func replayKeys(t *testing.T, client *redis.Client) []string {
	t.Helper()
	keys, err := client.Keys(context.Background(), "replay:*").Result()
	if err != nil {
		t.Fatalf("keys replay:*: %v", err)
	}
	return keys
}

func TestReplayGravaEntradaPorMudancaAplicada(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")

	room := newReplayRoom("sala-aplicada", "00000000-0000-7000-8000-000000000001", true)
	seedReplayRoom(t, client, room)

	err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
		r.Turn = 2
		return true
	}, redisFuncs.ReplayPlayHint{Actor: "p1", Payload: []byte(`{"type":"TAKE_COINS"}`)})
	if err != nil {
		t.Fatalf("WithRoom: %v", err)
	}

	entries, err := redisFuncs.ReadReplayEntries(ctx, client, room.MatchID)
	if err != nil {
		t.Fatalf("ReadReplayEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("esperava 1 entrada, veio %d", len(entries))
	}
	e := entries[0]
	if e.Type != "play" || !e.Applied || e.Player != "p1" {
		t.Fatalf("entrada inesperada: %+v", e)
	}
	if string(e.Payload) != `{"type":"TAKE_COINS"}` {
		t.Fatalf("payload = %q", e.Payload)
	}

	// Atomicidade: room e replay gravados juntos.
	if n, _ := client.Exists(ctx, "room:"+room.ID).Result(); n != 1 {
		t.Fatalf("chave room não gravada")
	}
	if n, _ := client.Exists(ctx, redisFuncs.ReplayListKey(room.MatchID)).Result(); n != 1 {
		t.Fatalf("chave replay não gravada")
	}
}

func TestReplayEntradaInicio(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")

	room := newReplayRoom("sala-inicio", "00000000-0000-7000-8000-000000000002", false)
	seedReplayRoom(t, client, room)

	err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
		r.StartTime = time.Now().UTC()
		return true
	})
	if err != nil {
		t.Fatalf("WithRoom: %v", err)
	}

	entries, err := redisFuncs.ReadReplayEntries(ctx, client, room.MatchID)
	if err != nil {
		t.Fatalf("ReadReplayEntries: %v", err)
	}
	if len(entries) != 1 || entries[0].Type != "start" {
		t.Fatalf("esperava entrada 0 = start, veio %+v", entries)
	}
}

func TestReplayTickNaoRodadoNaoGrava(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")

	room := newReplayRoom("sala-tick", "00000000-0000-7000-8000-000000000003", true)
	seedReplayRoom(t, client, room)

	err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CauseTick, func(r *rooms.Room) bool {
		r.Turn = 999 // muta, mas o tick "não rodou"
		return false
	})
	if err != nil {
		t.Fatalf("WithRoom: %v", err)
	}

	if keys := replayKeys(t, client); len(keys) != 0 {
		t.Fatalf("tick que não rodou gravou replay: %v", keys)
	}
}

func TestReplayJogadaRecusadaGravaEntrada(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")

	room := newReplayRoom("sala-recusada", "00000000-0000-7000-8000-000000000004", true)
	seedReplayRoom(t, client, room)

	err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
		return false // recusada
	}, redisFuncs.ReplayPlayHint{Actor: "p1", Payload: []byte(`{"type":"ASSASSIN"}`)})
	if err != nil {
		t.Fatalf("WithRoom: %v", err)
	}

	entries, err := redisFuncs.ReadReplayEntries(ctx, client, room.MatchID)
	if err != nil {
		t.Fatalf("ReadReplayEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("esperava 1 entrada, veio %d", len(entries))
	}
	if entries[0].Type != "play" || entries[0].Applied {
		t.Fatalf("jogada recusada deveria gravar applied=false, veio %+v", entries[0])
	}
}

func TestReplaySaveFalhouNaoGrava(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")
	t.Setenv("ROOM_LOCK_TTL_MS", "300")

	room := newReplayRoom("sala-falha", "00000000-0000-7000-8000-000000000005", true)
	seedReplayRoom(t, client, room)

	config.InstanceID = "pod-a"
	saveResult := make(chan error, 1)

	go func() {
		err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
			if err := client.Del(ctx, "lock:room:"+room.ID).Err(); err != nil {
				saveResult <- err
				return false
			}
			deadline := time.Now().Add(2 * time.Second)
			for time.Now().Before(deadline) {
				if r.Lock().Err() != nil {
					break
				}
				time.Sleep(5 * time.Millisecond)
			}
			return true
		}, redisFuncs.ReplayPlayHint{Actor: "p1", Payload: []byte(`{"type":"TAKE_COINS"}`)})
		saveResult <- err
	}()

	err := <-saveResult
	if err == nil {
		t.Fatal("WithRoom deveria falhar ao salvar após a perda do lock")
	}

	if keys := replayKeys(t, client); len(keys) != 0 {
		t.Fatalf("save falhou mas gravou replay: %v", keys)
	}
}

func TestReplayReinicioCoerente(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")

	room := newReplayRoom("sala-reinicio", "00000000-0000-7000-8000-000000000006", true)
	seedReplayRoom(t, client, room)

	// Duas chamadas separadas (como se fossem instâncias distintas do recorder).
	if err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
		r.Turn = 2
		return true
	}, redisFuncs.ReplayPlayHint{Actor: "p1", Payload: []byte(`{"type":"TAKE_COINS"}`)}); err != nil {
		t.Fatalf("WithRoom 1: %v", err)
	}
	if err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CauseTick, func(r *rooms.Room) bool {
		r.Turn = 3
		return true
	}); err != nil {
		t.Fatalf("WithRoom 2: %v", err)
	}

	entries, err := redisFuncs.ReadReplayEntries(ctx, client, room.MatchID)
	if err != nil {
		t.Fatalf("ReadReplayEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("esperava 2 entradas, veio %d", len(entries))
	}
	if entries[1].AtMs < entries[0].AtMs {
		t.Fatalf("entradas fora de ordem: %d < %d", entries[1].AtMs, entries[0].AtMs)
	}
}

func TestReplayNaoContemSemente(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")

	room := newReplayRoom("sala-semente", "00000000-0000-7000-8000-000000000007", true)
	seedReplayRoom(t, client, room)

	if err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
		return true
	}, redisFuncs.ReplayPlayHint{Actor: "p1", Payload: []byte(`{"type":"TAKE_COINS"}`)}); err != nil {
		t.Fatalf("WithRoom: %v", err)
	}

	seedHex := hex.EncodeToString(room.Seed[:])
	for _, key := range replayKeys(t, client) {
		vals, err := client.LRange(ctx, key, 0, -1).Result()
		if err != nil {
			t.Fatalf("lrange %s: %v", key, err)
		}
		for _, v := range vals {
			if hexContains(v, seedHex) {
				t.Fatalf("chave %s contém a semente", key)
			}
		}
	}
}

func hexContains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func TestFinalizeEnviaHwrDecodificavel(t *testing.T) {
	_, client := newReplayRedis(t)
	ctx := context.Background()
	t.Setenv("REPLAY_ENABLED", "true")

	room := newReplayRoom("sala-final", "00000000-0000-7000-8000-000000000008", true)
	seedReplayRoom(t, client, room)

	if err := redisFuncs.WithRoom(ctx, client, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
		r.Turn = 2
		return true
	}, redisFuncs.ReplayPlayHint{Actor: "p1", Payload: []byte(`{"type":"TAKE_COINS"}`)}); err != nil {
		t.Fatalf("WithRoom: %v", err)
	}

	room.GameOver = true
	room.WinnerID = "p1"

	var received []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("x-api-key ausente/incorreta")
		}
		received, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	backend := &BackendService{url: ts.URL, apiKey: "test-key", client: ts.Client()}

	if err := FinalizeReplay(ctx, backend, client, room); err != nil {
		t.Fatalf("FinalizeReplay: %v", err)
	}

	if len(received) == 0 {
		t.Fatal("backend não recebeu o .hwr")
	}
	var upload struct {
		Data      string   `json:"data"`
		RoomID    string   `json:"room_id"`
		PlayerIDs []string `json:"player_ids"`
	}
	if err := json.Unmarshal(received, &upload); err != nil {
		t.Fatalf("backend recebeu envelope inválido: %v", err)
	}
	if upload.RoomID != room.ID || len(upload.PlayerIDs) != len(room.Players) {
		t.Fatalf("envelope inesperado: room=%q players=%v", upload.RoomID, upload.PlayerIDs)
	}
	raw, err := base64.StdEncoding.DecodeString(upload.Data)
	if err != nil {
		t.Fatalf("data base64 inválido: %v", err)
	}
	if _, err := format.Decode(raw, nil); err != nil {
		t.Fatalf("backend recebeu .hwr inválido: %v", err)
	}

	// A lista foi limpa só depois do envio.
	if keys := replayKeys(t, client); len(keys) != 0 {
		t.Fatalf("lista replay não foi limpa: %v", keys)
	}
}
