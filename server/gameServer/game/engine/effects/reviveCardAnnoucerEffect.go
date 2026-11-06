package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ReviveCardAnnouncer(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o reviveCardPayload
	reviveCardPayload, ok := effect.Payload.(structs.ReviveCardPayload)
	if !ok {
		utils.LogError("payload inválido para ReviveCardPayload")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*reviveCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	// cria o payload
	newRevivePayload := structs.NewReviveCardPayload(string(effect.Cause), &targetPlayer.Id, reviveCardPayload.TargetCardIndex)

	// registra o evento de apresentação
	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventRevivedCard, newRevivePayload)); err != nil {
		return err
	}

	// registra o efeito de ganho de moedas
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectRevivedCard, effect.SourcePlayer, newRevivePayload)); err != nil {
		return err
	}

	return nil
}
