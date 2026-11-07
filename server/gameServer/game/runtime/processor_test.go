package runtime

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newProcessorRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

func seedExpiredRooms(t *testing.T, client *redis.Client, count int) []string {
	t.Helper()
	ctx := context.Background()
	base := time.Now().Add(-time.Hour).UnixMilli()
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = "sala-" + strconv.Itoa(i)
		if err := client.ZAdd(ctx, "rooms:timeouts", redis.Z{
			Score:  float64(base + int64(i)),
			Member: ids[i],
		}).Err(); err != nil {
			t.Fatalf("seed timeout: %v", err)
		}
	}
	return ids
}

func TestDrainExpiredRoomsProcessa20SalasEmUmTick(t *testing.T) {
	_, client := newProcessorRedis(t)
	ctx := context.Background()
	t.Setenv("PROCESSOR_BATCH_SIZE", "5")
	t.Setenv("PROCESSOR_TICK_BUDGET_MS", "1000")

	ids := seedExpiredRooms(t, client, 20)

	var processed []string
	n, err := drainExpiredRooms(ctx, client, func(roomID string) {
		processed = append(processed, roomID)
		client.ZRem(ctx, "rooms:timeouts", roomID)
	})
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	if n != len(ids) {
		t.Fatalf("1 tick processou %d salas, quer %d", n, len(ids))
	}

	if len(processed) != len(ids) {
		t.Fatalf("processadas %d salas, quer %d", len(processed), len(ids))
	}
	for i, id := range ids {
		if processed[i] != id {
			t.Fatalf("ordem[%d] = %s, quer %s", i, processed[i], id)
		}
	}

	remaining, err := client.ZCard(ctx, "rooms:timeouts").Result()
	if err != nil {
		t.Fatalf("zcard: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("fila ainda tem %d salas", remaining)
	}
}

func TestDrainExpiredRoomsRespeitaOrcamentoDoTick(t *testing.T) {
	_, client := newProcessorRedis(t)
	ctx := context.Background()
	t.Setenv("PROCESSOR_BATCH_SIZE", "5")
	t.Setenv("PROCESSOR_TICK_BUDGET_MS", "20")

	// Sala que se reagenda para "agora" toda vez: sem o orçamento, o drain nunca
	// devolveria. O teto rígido evita que uma regressão trave a suíte.
	seedExpiredRooms(t, client, 1)

	const hardCap = 500
	calls := 0
	start := time.Now()
	n, err := drainExpiredRooms(ctx, client, func(roomID string) {
		calls++
		if calls >= hardCap {
			client.ZRem(ctx, "rooms:timeouts", roomID)
			return
		}
		time.Sleep(time.Millisecond)
		client.ZAdd(ctx, "rooms:timeouts", redis.Z{Score: 0, Member: roomID})
	})
	if err != nil {
		t.Fatalf("drain: %v", err)
	}

	if n == 0 {
		t.Fatal("drain não processou nenhuma sala")
	}
	if calls >= hardCap {
		t.Fatalf("drain não parou pelo orçamento (processou %d salas)", calls)
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("drain levou %v, quer parar pelo orçamento de 20ms", elapsed)
	}
}

func TestDrainExpiredRoomsOrcamentoZeroMantemUmLote(t *testing.T) {
	_, client := newProcessorRedis(t)
	ctx := context.Background()
	t.Setenv("PROCESSOR_BATCH_SIZE", "5")
	t.Setenv("PROCESSOR_TICK_BUDGET_MS", "0")

	seedExpiredRooms(t, client, 10)

	n, err := drainExpiredRooms(ctx, client, func(roomID string) {
		client.ZRem(ctx, "rooms:timeouts", roomID)
	})
	if err != nil {
		t.Fatalf("drain: %v", err)
	}
	if n != 5 {
		t.Fatalf("processadas %d salas, quer 1 lote (5)", n)
	}
}
