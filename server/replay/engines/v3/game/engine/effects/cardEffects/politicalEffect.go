package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

func PoliticalEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	cardRules, err := roomData.GetCardRules(registryRules, string(effect.Cause))
	if err != nil {
		pure.LogError(err)
		return
	}

	// dobra o valor da carta específica
	roomData.MarkCardValueAsDoubled(registryRules, structs.TypePlayerPlays(effect.Cause))

	// incrementa a taxa da sala
	if cardRules.FixedValue != nil {
		roomData.IncrementTax(*cardRules.FixedValue)
	}

	// taxas maiores podem liberar mais investimentos
	roomData.EnforceInvestorMaxInvestments(registryRules)

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventIncrementTaxes, effect.Payload)); err != nil {
		pure.LogError(err)
		return
	}
}
