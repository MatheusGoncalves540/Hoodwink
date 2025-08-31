package validators

import (
	"context"

	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// return true if evt is a valid start event
func StartValidator(evt *rs.Event, ctx context.Context, rdb *redis.Client, room *rs.Room) bool {
	return true
}
