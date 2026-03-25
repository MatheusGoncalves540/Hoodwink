package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func KamikazeProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, kamikazePayload *structs.KamikazePayload) error {
	killedHimSelf := kamikazePayload.KilledHimSelf && kamikazePayload.TargetAllyCardIndex != nil

	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	if killedHimSelf {
		sourcePlayer.KillCard(*kamikazePayload.TargetAllyCardIndex)
	}

	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventCardPlayedKamikaze, expiresAt, kamikazePayload)
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectKamikaze, sourcePlayer.Id, kamikazePayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito kamikaze: %w", err)
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)

	return nil
}
