package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

func InvestmentEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	investmentPayload, ok := effect.Payload.(structs.InvestmentPayload)
	if !ok {
		pure.LogError("payload inválido para EffectInvestment")
		return
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		pure.LogError(err)
		return
	}

	// Aplica o investimento ao jogador
	sourcePlayer.AddInvestment(*investmentPayload.GracePeriod)
}
