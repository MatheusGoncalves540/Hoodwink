package redisFuncs

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// ScanKeys percorre o keyspace com SCAN (cursor + MATCH), em vez do KEYS que é
// O(N) e bloqueia o Redis. Devolve as chaves encontradas; o MATCH é o glob do Redis.
func ScanKeys(ctx context.Context, rdb *redis.Client, match string) ([]string, error) {
	var keys []string
	var cursor uint64

	for {
		page, next, err := rdb.Scan(ctx, cursor, match, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, page...)
		if next == 0 {
			return keys, nil
		}
		cursor = next
	}
}
