package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func RebelEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	cardRules, err := roomData.GetCardRules(registryRules, string(effect.Cause))
	if err != nil {
		utils.LogError(err)
		return
	}

	// dobra o valor da carta específica
	roomData.MarkCardValueAsDoubled(registryRules, structs.TypePlayerPlays(effect.Cause))

	// decrementa a taxa da sala
	roomData.DecrementTax(*cardRules.FixedValue)

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventDecrementTaxes, effect.Payload)); err != nil {
		utils.LogError(err)
		return
	}
}
