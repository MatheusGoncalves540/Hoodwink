package wsMsgHandler

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/wsMsgHandler/validators"
	"github.com/redis/go-redis/v9"
)

func OnStartAction(evt *roomStructs.Event, ctx context.Context, rdb *redis.Client, room *roomStructs.Room) error {
	if !validators.StartValidator(evt, ctx, rdb, room) {
		return fmt.Errorf("invalid request to start event")
	}
	// Start the playsProcessor goRoutine

	return nil
}
