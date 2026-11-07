package commands

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/redis/go-redis/v9"
)

func CommandSendBroadcast(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, payloadSent any) {
	redisFuncs.BroadcastMsgToRoom(ctx, rdb, roomData, payloadSent)
}
