package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

// BasicKillProtocol aplica a ação básica de pagar (6 + taxas/2) para matar uma carta.
// A Guardiã pode impedir o ataque.
func BasicKillProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, payload structs.AssassinPayload) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	price, err := roomData.GetBasicKillCardCost(registryRules)
	if err != nil {
		return err
	}
	sourcePlayer.RemoveCoins(price)

	killPayload := structs.NewKillCardPayload(string(structs.EffectBasicKill), payload.TargetPlayer, payload.TargetCardIndex)
	killEffect := structs.NewEffect(structs.EffectBasicKill, playerPlay.PlayerId, killPayload)

	if err := effects.KillAnnouncerCard(ctx, sink, registryRules, roomData, killEffect); err != nil {
		return err
	}

	lockAction(ctx, sink, roomData, playerPlay.PlayerId)
	return nil
}
