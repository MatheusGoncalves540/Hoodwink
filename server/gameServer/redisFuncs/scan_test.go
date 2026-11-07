package redisFuncs

import (
	"context"
	"sort"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newScanRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

func TestScanKeysEncontraTodasAsChaves(t *testing.T) {
	_, client := newScanRedis(t)
	ctx := context.Background()

	for _, id := range []string{"a", "b", "c", "d", "e"} {
		if err := client.Set(ctx, "player:"+id+":room", "r1:i1", 0).Err(); err != nil {
			t.Fatalf("set: %v", err)
		}
	}
	// Chave que não deve casar com o MATCH.
	if err := client.Set(ctx, "instance:i1:alive", "1", 0).Err(); err != nil {
		t.Fatalf("set: %v", err)
	}

	keys, err := ScanKeys(ctx, client, "player:*:room")
	if err != nil {
		t.Fatalf("ScanKeys: %v", err)
	}
	if len(keys) != 5 {
		t.Fatalf("esperado 5 chaves, veio %d: %v", len(keys), keys)
	}
}

func TestScanKeysVazio(t *testing.T) {
	_, client := newScanRedis(t)
	keys, err := ScanKeys(context.Background(), client, "player:*:room")
	if err != nil {
		t.Fatalf("ScanKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("esperado 0 chaves, veio %d", len(keys))
	}
}

func TestScanKeysPaginaVariasVezes(t *testing.T) {
	_, client := newScanRedis(t)
	ctx := context.Background()

	const total = 250
	for i := 0; i < total; i++ {
		id := string(rune('a'+i%26)) + string(rune('0'+i/26))
		if err := client.Set(ctx, "player:"+id+":room", "r1:i1", 0).Err(); err != nil {
			t.Fatalf("set: %v", err)
		}
	}

	keys, err := ScanKeys(ctx, client, "player:*:room")
	if err != nil {
		t.Fatalf("ScanKeys: %v", err)
	}
	sort.Strings(keys)
	if len(keys) != total {
		t.Fatalf("esperado %d chaves, veio %d", total, len(keys))
	}
	// Sem duplicatas mesmo atravessando múltiplas páginas do SCAN.
	seen := make(map[string]bool, len(keys))
	for _, k := range keys {
		if seen[k] {
			t.Fatalf("chave duplicada: %s", k)
		}
		seen[k] = true
	}
}
