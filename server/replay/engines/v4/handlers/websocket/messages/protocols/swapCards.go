package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

// SwapCardsProtocol aplica a ação básica de trocar cartas com o baralho.
// Uma carta custa metade das moedas (arredondada para cima); duas cartas custam todas as moedas.
func SwapCardsProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay, payload structs.SwapCardsPayload) error {
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

	if err := effects.ChangeCardAnnouncer(ctx, sink, registryRules, roomData, changeCardEffect); err != nil {
		return err
	}

	lockAction(ctx, sink, roomData, playerPlay.PlayerId)
	return nil
}
