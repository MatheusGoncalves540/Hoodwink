package redisHandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func SaveRoom(ctx context.Context, rdb *redis.Client, room *roomStructs.Room) error {
	instanceID := utils.GetInstanceID()
	ok, err := AcquireRoomLock(ctx, rdb, room.ID, instanceID, 5*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("a sala %s está sendo modificada por outra instância, tente novamente", room.ID)
	}
	defer ReleaseRoomLock(ctx, rdb, room.ID, instanceID)

	data, err := json.Marshal(room)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, "room:"+room.ID, data, 0).Err()
}
