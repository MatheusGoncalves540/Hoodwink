package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

func InvestorProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
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

	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		return err
	}
	gracePeriod := 6
	if cardRules.GracePeriod != nil {
		gracePeriod = *cardRules.GracePeriod
	}

	// Cada investimento custa 1 moeda da mão do jogador
	sourcePlayer.RemoveCoins(1)

	investorPayload := structs.NewInvestorPayload(&gracePeriod)

	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedInvestor, expiresAt, investorPayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectInvestor, playerPlay.PlayerId, investorPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito investor: %w", err)
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
	return nil
}
