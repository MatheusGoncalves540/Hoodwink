package engine

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// OnPlayerEvent processa quando uma ação é enviada por um jogador.
func OnPlayerEvent(
	ctx context.Context,
	rdb *redis.Client,
	roomID string,
	event roomStructs.PendingEvent,
) error {
	instanceID := utils.GetInstanceID()

	ok, err := room.AcquireRoomLock(ctx, rdb, roomID, instanceID, 2*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		// outra instância está processando a sala
		return nil
	}
	defer room.ReleaseRoomLock(ctx, rdb, roomID, instanceID)

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

	if err := room.SaveRoom(ctx, rdb, roomData); err != nil {
		return err
	}

	return nil
}
