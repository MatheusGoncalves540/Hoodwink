package wsMsgHandler

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/engineHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func ProcessDebugCommand(playerPlay *roomStructs.PendingEvent, ctx context.Context, rdb *redis.Client, roomId string) {
	switch playerPlay.Type {
	case "COMMAND_SEND_BROADCAST":
		engineHandlers.CommandSendBroadcast(ctx, rdb, roomId, playerPlay.Payload)
	case "COMMAND_PRINT_ROOM":
		engineHandlers.CommandPrintRoom(ctx, rdb, roomId)
	}
}
