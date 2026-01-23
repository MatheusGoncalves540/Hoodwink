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

func TrillionaireProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		return err
	}

	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	trillionairePayload := structs.NewTrillionairePayload(*cardRules.AmountReceived)

	roomData.GameEvent = structs.NewGameEvent(playerPlay.PlayerId, structs.EventCardPlayedTrillionaire, expiresAt, trillionairePayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectTrillionaire, playerPlay.PlayerId, trillionairePayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito trillionaire: %w", err)
	}

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
	return nil
}
