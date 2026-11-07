package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type heartbeatCommandCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (c *heartbeatCommandCounter) count(cmds ...redis.Cmder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cmd := range cmds {
		c.counts[cmd.Name()]++
	}
}

func (c *heartbeatCommandCounter) get(name string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[name]
}

func (c *heartbeatCommandCounter) DialHook(next redis.DialHook) redis.DialHook { return next }

func (c *heartbeatCommandCounter) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		c.count(cmd)
		return next(ctx, cmd)
	}
}

func (c *heartbeatCommandCounter) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		c.count(cmds...)
		return next(ctx, cmds)
	}
}

func TestSendHeartbeatRenovaSoPlayersDaInstanciaComSalasVivas(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	oldInstanceID := config.InstanceID
	config.InstanceID = "pod-a"
	t.Cleanup(func() { config.InstanceID = oldInstanceID })

	t.Setenv("HEARTBEAT_INTERVAL", "10")
	t.Setenv("HEARTBEAT_TTL", "30")

	for _, roomID := range []string{"sala-a", "sala-b"} {
		if err := client.Set(ctx, "room:"+roomID, "{}", 0).Err(); err != nil {
			t.Fatalf("seed sala: %v", err)
		}
	}
	players := map[string]string{
		"p1": "sala-a:pod-a",
		"p2": "sala-a:pod-a",
		"p3": "sala-b:pod-a",
		"p4": "sala-b:pod-b", // outra instância
		"p5": "sala-x:pod-a", // sala inexistente
	}
	for playerId, value := range players {
		if err := client.Set(ctx, "player:"+playerId+":room", value, time.Second).Err(); err != nil {
			t.Fatalf("seed player: %v", err)
		}
	}

	counter := &heartbeatCommandCounter{counts: make(map[string]int)}
	client.AddHook(counter)

	svc := NewHeartbeatService(client)
	svc.sendHeartbeat()

	// Instância viva e players com sala existente renovados para o TTL do heartbeat.
	if !mr.Exists("instance:pod-a:alive") {
		t.Fatal("heartbeat da instância não gravado")
	}
	for _, playerId := range []string{"p1", "p2", "p3"} {
		if ttl := mr.TTL("player:" + playerId + ":room"); ttl < 20*time.Second {
			t.Fatalf("TTL de %s = %v, quer ~30s", playerId, ttl)
		}
	}
	// Outra instância e sala inexistente não são tocadas.
	for _, playerId := range []string{"p4", "p5"} {
		if ttl := mr.TTL("player:" + playerId + ":room"); ttl > 5*time.Second {
			t.Fatalf("TTL de %s = %v, não deveria ter sido renovado", playerId, ttl)
		}
	}

	// Nenhum GET por jogador: os valores vêm de um único MGET.
	if got := counter.get("get"); got != 0 {
		t.Fatalf("sendHeartbeat fez %d GET(s), quer 0 (MGET/pipeline)", got)
	}
	if got := counter.get("mget"); got != 1 {
		t.Fatalf("sendHeartbeat fez %d MGET(s), quer 1", got)
	}
}
