package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/redis/go-redis/v9"
)

// OnPlayerEvent processa quando uma ação é enviada por um jogador.
func OnPlayerEvent(ctx context.Context, rdb *redis.Client, roomID string, event roomStructs.PendingEvent) error {
	if !room.TryLockRoom(ctx, rdb, roomID) {
		return nil
	}

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

	err = room.SaveRoom(ctx, rdb, roomData)
	if err != nil {
		return err
	}
	return nil
}
