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

func GravediggerProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, gravediggerPayload *structs.GravediggerPayload) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	// marca a remoção de coins
	cardPrice, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), 0)
	if err != nil {
		return err
	}

	sourcePlayer.RemoveCoins(cardPrice)

	// registra o evento de jogo
	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedGravedigger, expiresAt, gravediggerPayload)

	// registra o efeito pendente
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectGravedigger, playerPlay.PlayerId, gravediggerPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito gravedigger: %w", err)
	}

	// registra o timeout no redis
	roomData.RegistryTimeout(rdb, ctx, expiresAt)

	return nil
}
