package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisHandlers"
	"github.com/redis/go-redis/v9"
)

// OnPlayerEvent processa quando uma ação é enviada por um jogador.
func OnPlayerEvent(ctx context.Context, rdb *redis.Client, roomID string, event roomStructs.PendingEvent) error {
	if !redisHandlers.TryLockRoom(ctx, rdb, roomID) {
		return nil
	}

	room, err := redisHandlers.LoadRoom(ctx, rdb, roomID)
	if err != nil {
		return err
	}

	switch room.State {
	case roomStructs.StateWaitAction:
		// handleUseCard(room, event)

	case roomStructs.StateWaitContest:
		// handleContest(room, event)
	}

	err = redisHandlers.SaveRoom(ctx, rdb, room)
	if err != nil {
		return err
	}
	return nil
}
