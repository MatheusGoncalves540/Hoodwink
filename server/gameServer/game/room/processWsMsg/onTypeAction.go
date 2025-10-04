package processWsMsg

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/handlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/processWsMsg/validators"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func OnTypeAction(evt *rs.Event, ctx context.Context, rdb *redis.Client, room *rs.Room) error {
	payload, ok := evt.Payload.(map[string]any)
	if !ok {
		return fmt.Errorf("payload type assertion failed: %v", evt.Payload)
	}
	action, ok := payload["action"].(string)
	if !ok {
		return fmt.Errorf("action type assertion failed: %v", payload["action"])
	}

	switch action {
	case "use_assassin":
		if !validators.UseAssassinValidator(evt, ctx, rdb, room) {
			return fmt.Errorf("invalid request to use assassin")
		}
		if err := handlers.UseAssassin(ctx, rdb, room, evt); err != nil {
			return fmt.Errorf("error using assassin: %w", err)
		}
	default:
		return fmt.Errorf("unknown action: %v", action)
	}
	return nil
}
