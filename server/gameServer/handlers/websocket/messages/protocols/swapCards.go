package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

// SwapCardsProtocol aplica a ação básica de trocar cartas com o baralho.
// Uma carta custa metade das moedas (arredondada para cima); duas cartas custam todas as moedas.
func SwapCardsProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, payload structs.SwapCardsPayload) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	useOnTwoCards := payload.UseOnTwoCards != nil && *payload.UseOnTwoCards
	cost := roomData.GetBasicSwapCost(sourcePlayer.Coins, useOnTwoCards)
	sourcePlayer.RemoveCoins(cost)

	changeCardPayload := structs.NewChangeCardPayload(
		string(structs.EffectSwapCards),
		&sourcePlayer.Id,
		payload.TargetCardIndex,
		payload.UseOnTwoCards,
	)
	changeCardEffect := structs.NewEffect(structs.EffectSwapCards, playerPlay.PlayerId, changeCardPayload)

	if err := effects.ChangeCardAnnouncer(ctx, rdb, registryRules, roomData, changeCardEffect); err != nil {
		return err
	}

	lockAction(ctx, rdb, roomData, playerPlay.PlayerId)
	return nil
}
