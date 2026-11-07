package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

func KamikazeProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, kamikazePayload *structs.KamikazePayload) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := roomData.Now().Add(timeoutDuration * time.Second).UTC()

	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventCardPlayedKamikaze, expiresAt, kamikazePayload)
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectKamikaze, sourcePlayer.Id, kamikazePayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito kamikaze: %w", err)
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)

	return nil
}
