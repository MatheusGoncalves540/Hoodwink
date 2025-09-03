package wsRoom

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Publica mensagem para todas as instâncias
func PublishRoomBroadcast(ctx context.Context, rdb *redis.Client, roomId string, message []byte) error {
	return rdb.Publish(ctx, "room:"+roomId+":broadcast", message).Err()
}

// Assina o canal de broadcast de uma sala
func SubscribeRoomBroadcast(ctx context.Context, rdb *redis.Client, roomId string) {
	pubsub := rdb.Subscribe(ctx, "room:"+roomId+":broadcast")
	go func() {
		for msg := range pubsub.Channel() {
			// Quando receber mensagem do Redis, envia para os sockets locais
			ConnManager.Broadcast(roomId, []byte(msg.Payload))
		}
	}()
}
