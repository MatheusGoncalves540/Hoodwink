package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/rooms"
)

func ChangeCardAnnouncer(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o changeCardPayload
	changeCardPayload, ok := effect.Payload.(structs.ChangeCardPayload)
	if !ok {
		pure.LogError("payload inválido para ChangeCardPayload (announcer)")
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
