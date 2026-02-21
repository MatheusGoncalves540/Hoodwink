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

func GuardianProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, payload *structs.GuardianPayload) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "DisplayImportantMessage")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	guardianPayload := structs.GuardianPayload{}

	switch roomData.GameEvent.Type {
	case structs.EventWaitingFirstAction:
		guardianPayload = structs.NewGuardianPayload(payload.TargetPlayer, payload.TargetCardIndex, "future")

	case structs.EventCardKilled:
		guardianPayload = structs.NewGuardianPayload(payload.TargetPlayer, payload.TargetCardIndex, "death")

	case structs.EventCardPlayedClairvoyant:
		guardianPayload = structs.NewGuardianPayload(payload.TargetPlayer, payload.TargetCardIndex, "sight")
	}

	// se só tiver uma carta viva, o jogador pode usar Guardian por um desconto
	discount := 0
	aliveCards := sourcePlayer.GetAliveCardsIndexes()
	if len(aliveCards) == 1 {
		discount = 1
	}

	// marca a remoção de coins
	cardPrice, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), discount)
	if err != nil {
		return err
	}
	sourcePlayer.RemoveCoins(cardPrice)

	// registra o evento de jogo
	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedGuardian, expiresAt, guardianPayload)

	// registra o efeito pendente
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectGuardian, playerPlay.PlayerId, guardianPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito guardian: %w", err)
	}

	// registra o timeout no redis
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})

	// cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	// if err != nil {
	// 	return fmt.Errorf("%s", "Erro ao obter regras da carta Guardian: "+err.Error())
	// }

	// sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	// if err != nil {
	// 	return err
	// }

	// //

	// switch roomData.GameEvent.Type {
	// case structs.EventWaitingFirstAction:
	// 	// protege a carta requisitada
	// 	// fazer efeito para chamar  sourcePlayer.ProtectCard(payload.TargetCardIndex)
	// case structs.EventCardKilled:
	// 	// cancela a morte da carta requisitada
	// 	// fazer efeito para isso
	// }

	return nil
}
