package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func CroupierEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	croupierPayload, ok := effect.Payload.(structs.CroupierPayload)
	if !ok {
		pure.LogError("payload inválido para CroupierPayload")
		return
	}

	changeCardPayload := structs.NewChangeCardPayload(string(effect.Cause), croupierPayload.TargetPlayer, croupierPayload.TargetCardIndex, croupierPayload.UseOnTwoCards)

	changeCardEffect := structs.Effect{
		Cause:        structs.EffectChangedCard,
		SourcePlayer: effect.SourcePlayer,
		Payload:      changeCardPayload,
	}

	// chama o ChangeCardAnnouncer para trocar a carta do jogador alvo
	err := effects.ChangeCardAnnouncer(ctx, sink, registryRules, roomData, changeCardEffect)
	if err != nil {
		pure.LogError(err)
		return
	}
}
