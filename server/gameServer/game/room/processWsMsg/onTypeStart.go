package processWsMsg

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/processWsMsg/validators"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func OnStartAction(evt *rs.Event, ctx context.Context, rdb *redis.Client, room *rs.Room) error {
	if !validators.StartValidator(evt, ctx, rdb, room) {
		return fmt.Errorf("invalid request to start event")
	}
	// Start the playsProcessor goRoutine

	return nil
}
