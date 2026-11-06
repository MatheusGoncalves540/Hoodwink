package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

// BasicKillProtocol aplica a ação básica de pagar (6 + taxas/2) para matar uma carta.
// A Guardiã pode impedir o ataque.
func BasicKillProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, payload structs.AssassinPayload) error {
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

	if err := effects.KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect); err != nil {
		return err
	}

	lockAction(ctx, rdb, roomData, playerPlay.PlayerId)
	return nil
}
