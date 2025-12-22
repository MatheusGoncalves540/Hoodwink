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
func SubscribeRoomBroadcast(parentCtx context.Context, rdb *redis.Client, roomId string) {
	if ConnManager.RoomExists(roomId) {
		utils.LogDebug("Sala " + roomId + " já está assinada no Pub/Sub")
		return
	}

	ctx, cancel := context.WithCancel(parentCtx)
	ConnManager.SetRoomCancel(roomId, cancel)

	pubsub := rdb.Subscribe(ctx, "room:"+roomId+":broadcast")

	go func() {
		defer func() {
			_ = pubsub.Close()
			utils.LogDebug("Pub/Sub encerrado da sala " + roomId)
		}()

		ch := pubsub.Channel()

		for {
			select {
			case <-ctx.Done():
				return

			case msg, ok := <-ch:
				if !ok {
					return
				}
				ConnManager.Broadcast(roomId, []byte(msg.Payload))
			}
		}
	}()

	utils.LogDebug("Assinada sala " + roomId + " no Pub/Sub")
}
