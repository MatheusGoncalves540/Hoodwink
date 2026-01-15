package protocols

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func KamikazeProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, kamikazePayload *structs.KamikazePayload) error {
	killedHimSelf := kamikazePayload.KilledHimSelf && kamikazePayload.TargetAllyCardIndex != nil

	if killedHimSelf {
		sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
		if err != nil {
			return err
		}

		sourcePlayer.KillCard(*kamikazePayload.TargetAllyCardIndex)
	}

	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction") // TODO mudar tipo de timeout caso kamikaze esteja ativo na partida
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	roomData.GameEvent = &structs.GameEvent{
		PlayerID:  playerPlay.PlayerId,
		Type:      structs.EventCardPlayedKamikaze,
		ExpiresAt: expiresAt,
		Payload:   kamikazePayload,
	}
	roomData.PendingEffects = append(roomData.PendingEffects,
		structs.Effect{
			Cause:        structs.EffectKamikaze,
			SourcePlayer: playerPlay.PlayerId,
			Payload: structs.KamikazePayload{
				TargetPlayer:        kamikazePayload.TargetPlayer,
				TargetCardIndex:     kamikazePayload.TargetCardIndex,
				KilledHimSelf:       killedHimSelf,
				TargetAllyCardIndex: kamikazePayload.TargetAllyCardIndex,
			},
		},
	)

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})

	return nil
}
