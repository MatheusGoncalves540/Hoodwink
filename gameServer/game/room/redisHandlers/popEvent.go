package redisHandlers

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/redis/go-redis/v9"
)

// Remove e retorna o próximo evento da fila de eventos do Redis para a sala especificada.
// Usa BRPop para esperar até que um evento esteja disponível.
// Retorna o evento desempilhado ou nil se não houver evento.
// Em caso de erro de deserialização ou Redis, retorna o erro correspondente.
func PopEvent(ctx context.Context, rdb *redis.Client, roomID string) (*eventQueue.Event, error) {
	// Busca o próximo evento na fila (bloqueante até existir)
	res, err := rdb.BRPop(ctx, 0, "room:"+roomID+":eventQueue").Result()
	if err != nil {
		return nil, err // erro ao acessar o Redis
	}
	if len(res) < 2 {
		return nil, nil // nenhum evento disponível
	}
	var evt eventQueue.Event
	// Converte o JSON armazenado em struct Event
	err = json.Unmarshal([]byte(res[1]), &evt)
	if err != nil {
		return nil, err // erro ao desserializar o evento
	}
	return &evt, nil // retorna o evento desempilhado
}
