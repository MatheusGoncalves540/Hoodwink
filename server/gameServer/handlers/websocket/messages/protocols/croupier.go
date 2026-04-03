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

func CroupierProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, croupierPayload *structs.CroupierPayload) error {
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
	if croupierPayload.UseOnTwoCards != nil && *croupierPayload.UseOnTwoCards {
		// Se o croupier for usado em 2 cartas, remove todas as coins do jogador
		cardPrice := sourcePlayer.Coins
		sourcePlayer.RemoveCoins(cardPrice)
	} else {
		// Se o croupier for usado em 1 carta, remove metade das coins do jogador, arredondando para baixo
		cardPrice := sourcePlayer.Coins / 2
		sourcePlayer.RemoveCoins(cardPrice)
	}

	// registra o evento de jogo
	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedCroupier, expiresAt, croupierPayload)

	// registra o efeito pendente
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectCroupier, playerPlay.PlayerId, croupierPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito croupier: %w", err)
	}

	// registra o timeout no redis
	roomData.RegistryTimeout(rdb, ctx, expiresAt)

	return nil
}
