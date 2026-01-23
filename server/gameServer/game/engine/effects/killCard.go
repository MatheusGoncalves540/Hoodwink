package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects/effectsValidations"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

func KillCard(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	var payload structs.KillCardPayload
	switch typed := effect.Payload.(type) {
	case structs.KillCardPayload:
		payload = typed
	case *structs.KillCardPayload:
		payload = *typed
	default:
		if err := mapstructure.Decode(effect.Payload, &payload); err != nil {
			return err
		}
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

	// calcula o tempo de expiração do efeito
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction") // TODO mudar tipo de timeout caso kamikaze esteja ativo na partida
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	if err != nil {
		return err
	}

	// marca a targetCard do targetPlayer como morta
	err = targetPlayer.KillCard(*payload.TargetCardIndex)
	if err != nil {
		return err
	}

	killPayload := structs.NewKillCardPayload(string(effect.Cause), payload.TargetPlayer, payload.TargetCardIndex)

	// cria o evento pendente de carta morta
	roomData.GameEvent = structs.NewGameEvent(effect.SourcePlayer, structs.EventCardKilled, expiresAt, killPayload)
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
	return nil
}
