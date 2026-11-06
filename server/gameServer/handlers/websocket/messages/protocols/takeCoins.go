package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// TakeCoinsProtocol aplica a ação básica de coletar 1 moeda + taxas.
// Nada pode impedir essa ação.
func TakeCoinsProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	amount, err := roomData.GetBasicTakeCoinsAmount(registryRules)
	if err != nil {
		return err
	}

	effect := structs.NewEffect(
		structs.EffectTakeCoins,
		playerPlay.PlayerId,
		structs.NewEarnCoinsPayload(string(structs.EffectTakeCoins), amount, nil),
	)

	if err := effects.EarnCoinsAnnouncer(ctx, rdb, roomData, effect); err != nil {
		utils.LogError(err)
		return err
	}

	lockAction(ctx, rdb, roomData, playerPlay.PlayerId)
	return nil
}
