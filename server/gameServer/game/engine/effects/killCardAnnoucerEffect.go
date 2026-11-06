package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects/effectsValidations"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func KillAnnouncerCard(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o killCardPayload
	killCardPayload, ok := effect.Payload.(structs.KillCardPayload)
	if !ok {
		utils.LogError("payload inválido para KillCardPayload")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*killCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	// valida o efeito
	valid, err := effectsValidations.ValidateKillCardEffect(roomData, effect, killCardPayload, targetPlayer)
	if err != nil || !valid {
		return err
	}

	newKillPayload := structs.NewKillCardPayload(string(effect.Cause), killCardPayload.TargetPlayer, killCardPayload.TargetCardIndex)

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventCardKilled, newKillPayload)); err != nil {
		return err
	}

	// registra o efeito de ganho de moedas
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectCardKilled, effect.SourcePlayer, newKillPayload)); err != nil {
		return err
	}

	return nil
}
