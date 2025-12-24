package messages

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/redis/go-redis/v9"
)

func ProcessPlay(ctx context.Context, rdb *redis.Client, roomID string, playerPlay *roomStructs.PendingEvent) error {
	ok, err := room.AcquireRoomLock(ctx, rdb, roomID, config.InstanceID, 2*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	defer room.ReleaseRoomLock(ctx, rdb, roomID, config.InstanceID)

	roomData, err := room.LoadRoom(ctx, rdb, roomID)
	if err != nil {
		return err
	}

	switch roomData.State {
	case roomStructs.StateWaitAction:
		// handleUseCard(roomData, event)

	case roomStructs.StateWaitContest:
		// handleContest(roomData, event)
	}
	return nil
}
