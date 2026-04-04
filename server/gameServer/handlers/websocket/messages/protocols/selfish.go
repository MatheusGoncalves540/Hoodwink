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
		return fmt.Errorf("%s", "Erro ao obter regras da carta Rebel: "+err.Error())
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

	// marca a remoção de coins
	switch roomData.GameEvent.Type {
	case structs.EventChangedCard:
		cardPrice, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), 0)
		if err != nil {
			return err
		}

		sourcePlayer.RemoveCoins(cardPrice)
	case structs.EventEarnCoins:
		cardPrice := *cardRules.FixedValue
		sourcePlayer.RemoveCoins(cardPrice)

	case structs.EventInvestment:
		// disconto de 5, pra contar só o fix de 1 e as taxas
		cardPrice, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), 5)
		if err != nil {
			return err
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
