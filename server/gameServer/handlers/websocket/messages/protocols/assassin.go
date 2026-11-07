package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func AssassinProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, assassinPayload structs.AssassinPayload) error {
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

	sourcePlayer.RemoveCoins(cardPrice)

	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventCardPlayedAssassin, expiresAt, assassinPayload)
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectAssassin, sourcePlayer.Id, assassinPayload)); err != nil {
		pure.LogError(err)
		return fmt.Errorf("falha ao registrar efeito assassin: %w", err)
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
	return nil
}
