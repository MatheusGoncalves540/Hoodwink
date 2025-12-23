package wsMsgHandler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisHandlers"
	"github.com/redis/go-redis/v9"
)

func ProcessDebugCommand(playerPlay *roomStructs.PendingEvent, ctx context.Context, rdb *redis.Client, roomId string) {
	switch playerPlay.Type {
	case "COMMAND_SEND_BROADCAST":
		wsRoom.PublishRoomBroadcast(ctx, rdb, roomId, playerPlay.Payload)
	case "COMMAND_PRINT_ROOM":
		room, err := redisHandlers.LoadRoom(ctx, rdb, roomId)
		if err != nil {
			println("Erro ao obter sala:", err.Error())
			return
		}
		data, _ := json.MarshalIndent(room, "", "  ")
		fmt.Println(string(data))
	}
}
