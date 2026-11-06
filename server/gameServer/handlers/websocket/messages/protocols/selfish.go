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

func SelfishProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return fmt.Errorf("falha ao obter jogador: %w", err)
	}

	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		return fmt.Errorf("erro ao obter regras da carta Selfish: %w", err)
	}

	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	// salva o que foi negado ao usar Selfish
	whatWasDeniedValue := string(roomData.GameEvent.Type)
	whatWasDenied := &whatWasDeniedValue

	// salva o player que foi impedido de ganhar moedas, investimento ou troca de carta
	targetPlayerValue := roomData.GameEvent.PlayerID
	targetPlayer := &targetPlayerValue

	// marca a remoção de coins de acordo com o que está sendo negado
	switch roomData.GameEvent.Type {
	case structs.EventChangedCard, structs.EventRevivedCard:
		cardPrice, err := roomData.GetCardValue(registryRules, playerPlay.Type, 0)
		if err != nil {
			return err
		}
		sourcePlayer.RemoveCoins(cardPrice)

	default:
		// negar moedas, roubo ou investimento custa 1 moeda fixa
		cardPrice := 1
		if cardRules.FixedValue != nil {
			cardPrice = *cardRules.FixedValue
		}
		sourcePlayer.RemoveCoins(cardPrice)
	}

	selfishPayload := structs.NewSelfishPayload(whatWasDenied, targetPlayer)

	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedSelfish, expiresAt, selfishPayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectSelfish, playerPlay.PlayerId, selfishPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito selfish: %w", err)
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
	return nil
}
