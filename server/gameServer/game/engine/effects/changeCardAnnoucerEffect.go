package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ChangeCardAnnouncer(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o changeCardPayload
	changeCardPayload, ok := effect.Payload.(structs.ChangeCardPayload)
	if !ok {
		utils.LogError("payload inválido para ChangeCardPayload (announcer)")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*changeCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	// cria o payload
	newChangePayload := structs.NewChangeCardPayload(string(effect.Cause), &targetPlayer.Id, changeCardPayload.TargetCardIndex, changeCardPayload.UseOnTwoCards)

	// registra o evento de apresentação
	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventChangedCard, newChangePayload)); err != nil {
		return err
	}

	// registra o efeito de troca de carta
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectChangedCard, effect.SourcePlayer, newChangePayload)); err != nil {
		return err
	}

	return nil
}
