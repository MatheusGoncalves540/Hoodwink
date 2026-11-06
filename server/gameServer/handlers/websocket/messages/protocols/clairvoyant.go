package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ClairvoyantProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, clairvoyantPayload structs.ClairvoyantPayload) error {
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

	// remove coins from source player
	sourcePlayer.RemoveCoins(cardPrice)

	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventCardPlayedClairvoyant, expiresAt, clairvoyantPayload)
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectClairvoyant, sourcePlayer.Id, clairvoyantPayload)); err != nil {
		utils.LogError(err)
		return fmt.Errorf("falha ao registrar efeito clairvoyant: %w", err)
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
	return nil
}
