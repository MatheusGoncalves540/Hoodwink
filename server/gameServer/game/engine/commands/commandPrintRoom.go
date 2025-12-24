package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/redis/go-redis/v9"
)

func CommandPrintRoom(ctx context.Context, rdb *redis.Client, roomId string) {
	roomData, err := room.LoadRoom(ctx, rdb, roomId)
	if err != nil {
		println("Erro ao obter sala:", err.Error())
		return
	}
	data, _ := json.MarshalIndent(roomData, "", "  ")
	fmt.Println(string(data))
}
