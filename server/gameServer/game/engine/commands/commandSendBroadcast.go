package commands

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/redis/go-redis/v9"
)

func CommandSendBroadcast(ctx context.Context, rdb *redis.Client, roomId string, payloadSent any) {
	wsRoom.PublishRoomBroadcast(ctx, rdb, roomId, payloadSent)
}
