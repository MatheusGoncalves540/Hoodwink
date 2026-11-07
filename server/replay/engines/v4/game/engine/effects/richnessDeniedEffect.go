package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

func RichnessDeniedEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	deniedPayload, ok := effect.Payload.(structs.RichnessDeniedPayload)
	if !ok {
		pure.LogError("payload inválido para RichnessDeniedPayload")
		return
	}

	// Se o que foi negado foi um investimento, o efeito de investimento é cancelado
	// e a moeda usada para investir volta para a mão do jogador.
	lastEffect, hasLast := roomData.PeekLastPendingEffect()
	if hasLast && lastEffect.Cause == structs.EffectInvestment && deniedPayload.TargetPlayer != nil {
		targetPlayer, err := roomData.GetPlayer(*deniedPayload.TargetPlayer)
		if err != nil {
			pure.LogError(err)
		} else if generalRules, err := roomData.GetGeneralRules(registryRules); err != nil {
			pure.LogError(err)
		} else {
			targetPlayer.AddCoins(1, *generalRules.MaxCoins)
		}
	}

	roomData.PopLastPendingEffect()
}
