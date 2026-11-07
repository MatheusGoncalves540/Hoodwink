package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func ClairvoyantProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, clairvoyantPayload structs.ClairvoyantPayload) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := roomData.Now().Add(timeoutDuration * time.Second).UTC()

	// marca a remoção de coins
	cardPrice, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), 0)
	if err != nil {
		return err
	}

	// remove coins from source player
	sourcePlayer.RemoveCoins(cardPrice)

	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventCardPlayedClairvoyant, expiresAt, clairvoyantPayload)
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectClairvoyant, sourcePlayer.Id, clairvoyantPayload)); err != nil {
		pure.LogError(err)
		return fmt.Errorf("falha ao registrar efeito clairvoyant: %w", err)
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
	return nil
}
