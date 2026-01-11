package commands

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/redis/go-redis/v9"
)

func CommandSendBroadcast(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, payloadSent any) {
	roomData.PublishRoomBroadcast(ctx, rdb, payloadSent)
}
