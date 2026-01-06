package protocols

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/redis/go-redis/v9"
)

func AssassinProtocol(ctx context.Context, rdb *redis.Client, roomData *roomStructs.Room, playerPlay *roomStructs.PlayerPlay, assassinPayload roomStructs.AssassinPayload) error {

	expiresAt := time.Now().Add(7 * time.Second).UTC()
	roomData.GameEvent = &roomStructs.GameEvent{
		PlayerID:  playerPlay.PlayerId,
		Type:      roomStructs.EventCardPlayedAssassin,
		ExpiresAt: expiresAt, // TODO colocar tempo configuravel
		Payload:   assassinPayload,
	}
	roomData.PendingEffects = append(roomData.PendingEffects,
		roomStructs.Effect{
			Cause:        roomStructs.EffectAssassin,
			SourcePlayer: playerPlay.PlayerId,
			Payload: roomStructs.AssassinPayload{
				TargetPlayer:    assassinPayload.TargetPlayer,
				TargetCardIndex: assassinPayload.TargetCardIndex,
			},
		},
	)

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
	room.SaveRoom(ctx, rdb, roomData)
	wsRoom.PublishRoomBroadcast(ctx, rdb, roomData.ID, roomData)
	return nil
}
