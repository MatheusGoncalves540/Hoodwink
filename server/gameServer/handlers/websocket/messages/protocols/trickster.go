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

func TricksterProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, tricksterPayload *structs.TricksterPayload) error {
	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	// calcula as coins roubadas
	stealedCoins, err := roomData.GetCardValue(registryRules, structs.TypePlayerPlays(playerPlay.Type), 0)
	if err != nil {
		return err
	}
	tricksterPayload.StealedCoins = stealedCoins

	// cria o evento de jogo
	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedTrickster, expiresAt, tricksterPayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectTrickster, playerPlay.PlayerId, tricksterPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito trickster: %w", err)
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
	return nil
}
