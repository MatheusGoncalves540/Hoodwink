package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

func TrillionaireProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := roomData.Now().Add(timeoutDuration * time.Second).UTC()

	// calcula as coins ganhas
	earnedCoins, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), 0)
	if err != nil {
		return err
	}

	trillionairePayload := structs.NewTrillionairePayload(earnedCoins, nil)

	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedTrillionaire, expiresAt, trillionairePayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectTrillionaire, playerPlay.PlayerId, trillionairePayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito trillionaire: %w", err)
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
	return nil
}
