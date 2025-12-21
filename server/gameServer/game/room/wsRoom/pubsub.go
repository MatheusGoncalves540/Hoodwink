package wsRoom

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

// Publica mensagem para todos os players de uma sala
func PublishRoomBroadcast(ctx context.Context, rdb *redis.Client, roomId string, message any) error {
	if pubMsg, err := json.Marshal(message); err == nil {
		return rdb.Publish(ctx, "room:"+roomId+":broadcast", pubMsg).Err()
	}
	return nil
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
