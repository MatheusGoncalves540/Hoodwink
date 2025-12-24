package messages

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func ProcessPlay(playerPlay *roomStructs.PendingEvent, ctx context.Context, rdb *redis.Client) {

}
