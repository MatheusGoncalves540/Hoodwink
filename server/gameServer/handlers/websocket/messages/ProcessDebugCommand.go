package messages

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/commands"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func ProcessDebugCommand(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay) {
	switch playerPlay.Type {
	case "COMMAND_SEND_BROADCAST":
		commands.CommandSendBroadcast(ctx, rdb, roomData, playerPlay.Payload)
	case "COMMAND_PRINT_ROOM":
		commands.CommandPrintRoom(ctx, rdb, roomData)
	case "COMMAND_START_GAME":
		commands.CommandStartGame(ctx, rdb, roomData)
	}
}
