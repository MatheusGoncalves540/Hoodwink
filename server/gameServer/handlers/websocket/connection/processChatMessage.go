package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/runtime/chat"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/redis/go-redis/v9"
)

// ProcessChatMessage processa mensagens de chat e devolve aplicado. Não persiste
// (o WithRoom salva); apenas anexa e transmite.
func ProcessChatMessage(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	var chatHandlers = map[structs.TypePlayerPlays]func() bool{
		"CHAT_MESSAGE": func() bool {
			return chat.ProcessChatMessage(roomData, playerPlay)
		},
	}

	if handler, ok := chatHandlers[playerPlay.Type]; ok {
		if aplicado := handler(); aplicado {
			if err := redisFuncs.SendUpdatedRoomData(ctx, rdb, roomData, nil, []string{}); err != nil {
				return false
			}
			return true
		}
	}
	return false
}
