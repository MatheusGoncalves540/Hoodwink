package messages

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/chat"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func ProcessChatMessage(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay) {
	var chatHandlers = map[structs.TypePlayerPlays]func(){
		"CHAT_MESSAGE": func() {
			chat.ProcessChatMessage(ctx, rdb, roomData, playerPlay)
		},
	}

	if handler, ok := chatHandlers[playerPlay.Type]; ok {
		handler()
		if err := roomData.SaveRoom(ctx, rdb); err != nil {
			return
		}
		if err := roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{}); err != nil {
			return
		}
	}
}
