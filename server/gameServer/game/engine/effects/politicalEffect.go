package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func PoliticalEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	cardRules, err := roomData.GetCardRules(registryRules, string(effect.Cause))
	if err != nil {
		utils.LogError(err)
		return
	}

	// calcula o tempo de expiração do efeito
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "DisplayMessage")
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	if err != nil {
		utils.LogError(err)
		return
	}

	// dobra o valor da carta específica
	roomData.MarkCardValueAsDoubled(registryRules, structs.TypePlayerPlays(effect.Cause))

	// incrementa a taxa da sala
	roomData.IncrementTax(*cardRules.FixedValue)

	// cria o evento pendente de carta morta
	roomData.GameEvent = structs.NewGameEvent(effect.SourcePlayer, structs.EventIncrementTaxes, expiresAt, effect.Payload)
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
}
