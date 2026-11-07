package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

// TakeCoinsProtocol aplica a ação básica de coletar 1 moeda + taxas.
// Nada pode impedir essa ação.
func TakeCoinsProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	amount, err := roomData.GetBasicTakeCoinsAmount(registryRules)
	if err != nil {
		return err
	}

	effect := structs.NewEffect(
		structs.EffectTakeCoins,
		playerPlay.PlayerId,
		structs.NewEarnCoinsPayload(string(structs.EffectTakeCoins), amount, nil),
	)

	if err := effects.EarnCoinsAnnouncer(ctx, sink, roomData, effect); err != nil {
		pure.LogError(err)
		return err
	}

	lockAction(ctx, sink, roomData, playerPlay.PlayerId)
	return nil
}
