package commands

import (
	"context"
	"encoding/json"
	"log"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func CommandPrintRoom(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) {
	data, _ := json.MarshalIndent(roomData, "", "  ")
	log.Println(string(data))
}
