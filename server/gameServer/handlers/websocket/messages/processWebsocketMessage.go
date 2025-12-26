package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/redis/go-redis/v9"
)

func ProcessPlay(ctx context.Context, rdb *redis.Client, roomID string, playerPlay *roomStructs.PlayerPlay) error {
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

	switch playerPlay.Type {
	case roomStructs.PlayAssassinCard:
		// TODO Validar se o payload tem a estrutura de AssassinPayload de um jeito bom (agora ta provisório)
		var assassinPayload roomStructs.AssassinPayload
		payloadJSON, err := json.Marshal(playerPlay.Payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		if err := json.Unmarshal(payloadJSON, &assassinPayload); err != nil {
			return fmt.Errorf("payload does not match AssassinPayload structure: %w", err)
		}

		expiresAt := time.Now().Add(7 * time.Second).UTC()
		roomData.PendingEvent = &roomStructs.PendingEvent{
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
					TargetPlayer: assassinPayload.TargetPlayer,
					TargetCard:   assassinPayload.TargetCard,
				},
			},
		)

		rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
			Score:  float64(expiresAt.UnixMilli()),
			Member: roomData.ID,
		})
		room.SaveRoom(ctx, rdb, roomData)
		wsRoom.PublishRoomBroadcast(ctx, rdb, roomData.ID, roomData)
	}

	return nil
}
