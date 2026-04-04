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

func InvestorProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	// Calcula o período de carência para o investimento
	gracePeriod, err := roomData.GetCardValue(registryRules, playerPlay.Type, 0)
	if err != nil {
		return err
	}

	investorPayload := structs.NewInvestorPayload(&gracePeriod)

	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedInvestor, expiresAt, investorPayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectInvestor, playerPlay.PlayerId, investorPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito investor: %w", err)
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
	return nil
}
