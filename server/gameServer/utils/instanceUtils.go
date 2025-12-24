package utils

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/redis/go-redis/v9"
)

// LogInstanceInfo registra informações sobre a instância atual
func LogInstanceInfo(action string) {
	LogDebug(fmt.Sprintf("[Instance %s] %s", config.InstanceID, action))
}

// CleanupPlayerRegistrations remove registros de players órfãos (instâncias mortas)
func CleanupPlayerRegistrations(ctx context.Context, rdb *redis.Client) error {
	// Busca todas as chaves de player:*:room
	keys, err := rdb.Keys(ctx, "player:*:room").Result()
	if err != nil {
		return err
	}

	cleaned := 0
	for _, key := range keys {
		value, err := rdb.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		// Parse do formato "roomId:instanceId"
		parts := value
		lastColon := -1
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == ':' {
				lastColon = i
				break
			}
		}

		if lastColon == -1 {
			// Formato inválido, remove
			rdb.Del(ctx, key)
			cleaned++
			continue
		}

		instanceId := parts[lastColon+1:]

		// Verifica se a instância ainda está viva
		instanceKey := fmt.Sprintf("instance:%s:alive", instanceId)
		_, err = rdb.Get(ctx, instanceKey).Result()
		if err != nil {
			// Instância morta, remove registro do player
			rdb.Del(ctx, key)
			cleaned++
		}
	}

	if cleaned > 0 {
		LogDebug(fmt.Sprintf("Limpeza concluída: %d registros de players órfãos removidos", cleaned))
	}

	return nil
}

// Gera um UUID para identificar a instância/processo
func GetInstanceID() string {
	return config.InstanceID
}
