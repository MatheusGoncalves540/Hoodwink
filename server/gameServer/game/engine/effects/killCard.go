package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects/effectsValidations"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

func KillCard(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *roomStructs.Room, effect roomStructs.Effect) error {
	// decodifica o payload
	var payload roomStructs.KillCardPayload
	if err := mapstructure.Decode(effect.Payload, &payload); err != nil {
		return err
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*payload.TargetPlayer)
	if err != nil {
		return err
	}

	// valida o efeito
	valid, err := effectsValidations.ValidateKillCardEffect(roomData, effect, payload, targetPlayer)
	if err != nil || !valid {
		return err
	}

	// cria o evento pendente de carta morta
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "DisplayMessage") // TODO mudar tipo de timeout caso kamikaze esteja ativo na partida
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	if err != nil {
		return err
	}

	// marca a targetCard do targetPlayer como morta
	err = targetPlayer.KillCard(*payload.TargetCardIndex)
	if err != nil {
		return err
	}

	roomData.GameEvent = &roomStructs.GameEvent{
		PlayerID:  effect.SourcePlayer,
		Type:      roomStructs.EventCardKilled,
		ExpiresAt: expiresAt,
		Payload: map[string]interface{}{
			"TargetPlayer": *payload.TargetPlayer,
			"TargetCard":   *payload.TargetCardIndex,
			"Cause":        effect.Cause,
		},
	}
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
	return nil
}
