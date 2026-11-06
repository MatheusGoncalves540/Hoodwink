package messages

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/chat/commands"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func ProcessDebugCommand(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay) {
	var debugCommands = map[structs.TypePlayerPlays]func(){
		"COMMAND_SEND_BROADCAST": func() {
			commands.CommandSendBroadcast(ctx, rdb, roomData, playerPlay.Payload)
		},
		"COMMAND_PRINT_ROOM": func() {
			commands.CommandPrintRoom(ctx, rdb, roomData)
		},
		"COMMAND_START_GAME": func() {
			commands.CommandStartGame(ctx, rdb, roomData)
		},
	}

	if cmd, ok := debugCommands[playerPlay.Type]; ok {
		cmd()
		if err := roomData.SaveRoom(ctx, rdb); err != nil {
			return
		}
		if err := roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{}); err != nil {
			return
		}
	}
}
