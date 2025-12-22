package wsRoom

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
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
	if ConnManager.RoomExists(roomId) == true {
		utils.LogDebug("Sala " + roomId + " já está assinada no Pub/Sub")
		return
	}
	pubsub := rdb.Subscribe(ctx, "room:"+roomId+":broadcast")
	go func() {
		for msg := range pubsub.Channel() {
			// Quando receber mensagem do Redis, envia para os sockets locais
			ConnManager.Broadcast(roomId, []byte(msg.Payload))
		}
	}()
	utils.LogDebug("Assinada sala " + roomId + " no Pub/Sub")
}

// Cancela a assinatura do canal de broadcast de uma sala
func UnsubscribeRoomBroadcast(ctx context.Context, rdb *redis.Client, roomId string) {
	pubsub := rdb.Subscribe(ctx, "room:"+roomId+":broadcast")
	pubsub.Close()
	utils.LogDebug("Cancelada assinatura da sala " + roomId + " no Pub/Sub")
}
