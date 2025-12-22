package wsRoom

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/redis/go-redis/v9"
)

var subscribedRooms sync.Map // map[string]bool
var cancelFuncs sync.Map     // map[string]context.CancelFunc

// Publica mensagem para todos os players de uma sala
func PublishRoomBroadcast(ctx context.Context, rdb *redis.Client, roomId string, message any) error {
	if pubMsg, err := json.Marshal(message); err == nil {
		return rdb.Publish(ctx, "room:"+roomId+":broadcast", pubMsg).Err()
	}
	return nil
}

// Assina o canal de broadcast de uma sala
func SubscribeRoomBroadcast(ctx context.Context, rdb *redis.Client, roomId string) {
	if _, loaded := subscribedRooms.LoadOrStore(roomId, true); loaded {
		return // já inscrito
	}
	subCtx, cancel := context.WithCancel(context.Background())
	cancelFuncs.Store(roomId, cancel)
	pubsub := rdb.Subscribe(subCtx, "room:"+roomId+":broadcast")
	go func() {
		defer func() {
			cancelFuncs.Delete(roomId)
			subscribedRooms.Delete(roomId)
		}()
		for msg := range pubsub.Channel() {
			select {
			case <-subCtx.Done():
				return
			default:
				// Quando receber mensagem do Redis, envia para os sockets locais
				ConnManager.Broadcast(roomId, []byte(msg.Payload))
			}
		}
	}()
}

// UnsubscribeRoomBroadcast cancela a inscrição do pubsub para a room
func UnsubscribeRoomBroadcast(roomId string) {
	if cancelAny, ok := cancelFuncs.Load(roomId); ok {
		if cancel, ok2 := cancelAny.(context.CancelFunc); ok2 {
			cancel()
		}
		cancelFuncs.Delete(roomId)
	}
	subscribedRooms.Delete(roomId)
}
