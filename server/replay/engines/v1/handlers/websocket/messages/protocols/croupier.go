package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func CroupierProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, croupierPayload *structs.CroupierPayload) error {
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

	// Croupier custa (6 + taxas), com mínimo definido nas regras
	cardPrice, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), 0)
	if err != nil {
		return err
	}
	sourcePlayer.RemoveCoins(cardPrice)

	// registra o evento de jogo
	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedCroupier, expiresAt, croupierPayload)

	// registra o efeito pendente
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectCroupier, playerPlay.PlayerId, croupierPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito croupier: %w", err)
	}

	// registra o timeout no redis
	sink.RegistryTimeout(ctx, roomData, expiresAt)

	return nil
}
