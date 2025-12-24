package messages

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/commands"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func ProcessDebugCommand(playerPlay *roomStructs.PendingEvent, ctx context.Context, rdb *redis.Client, roomId string) {
	switch playerPlay.Type {
	case "COMMAND_SEND_BROADCAST":
		commands.CommandSendBroadcast(ctx, rdb, roomId, playerPlay.Payload)
	case "COMMAND_PRINT_ROOM":
		commands.CommandPrintRoom(ctx, rdb, roomId)
	}
}
