package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

func ReviveCardEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	reviveCardPayload, ok := effect.Payload.(structs.ReviveCardPayload)
	if !ok {
		pure.LogError("payload inválido para ReviveCardPayload")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*reviveCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	if reviveCardPayload.TargetCardIndex == nil {
		return nil
	}

	card, err := targetPlayer.GetCardByIndex(*reviveCardPayload.TargetCardIndex)
	if err != nil {
		return err
	}
	if !card.Dead {
		return nil
	}

	// revive a targetCard do targetPlayer
	err = targetPlayer.ReviveCard(*reviveCardPayload.TargetCardIndex)
	if err != nil {
		return err
	}

	return nil
}
