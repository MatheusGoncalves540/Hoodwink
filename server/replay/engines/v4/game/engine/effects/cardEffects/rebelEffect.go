package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

func RebelEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	cardRules, err := roomData.GetCardRules(registryRules, string(effect.Cause))
	if err != nil {
		pure.LogError(err)
		return
	}

	// dobra o valor da carta específica
	roomData.MarkCardValueAsDoubled(registryRules, structs.TypePlayerPlays(effect.Cause))

	// decrementa a taxa da sala
	if cardRules.FixedValue != nil {
		roomData.DecrementTax(*cardRules.FixedValue)
	}

	// taxas menores reduzem o limite de investimentos e podem descartar o excedente
	roomData.EnforceInvestorMaxInvestments(registryRules)

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventDecrementTaxes, effect.Payload)); err != nil {
		pure.LogError(err)
		return
	}
}
