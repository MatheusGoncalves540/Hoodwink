package wsMsgHandler

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/redis/go-redis/v9"
)

func ProcessDebugCommand(playerPlay *roomStructs.PlayerPlay, ctx context.Context, rdb *redis.Client) {
	switch playerPlay.Type {
	case "COMMAND_SEND_BROADCAST":
		wsRoom.PublishRoomBroadcast(ctx, rdb, playerPlay.RoomId, playerPlay.Data)
	}
}
