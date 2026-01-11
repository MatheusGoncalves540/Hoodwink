package messages

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/commands"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/redis/go-redis/v9"
)

func ProcessDebugCommand(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *roomStructs.PlayerPlay) {
	switch playerPlay.Type {
	case "COMMAND_SEND_BROADCAST":
		commands.CommandSendBroadcast(ctx, rdb, roomData, playerPlay.Payload)
	case "COMMAND_PRINT_ROOM":
		commands.CommandPrintRoom(ctx, rdb, roomData)
	}
}
