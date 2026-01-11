package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/redis/go-redis/v9"
)

func CommandPrintRoom(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) {
	data, _ := json.MarshalIndent(roomData, "", "  ")
	fmt.Println(string(data))
}
