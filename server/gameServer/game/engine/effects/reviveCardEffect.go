package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ReviveCardEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	reviveCardPayload, ok := effect.Payload.(structs.ReviveCardPayload)
	if !ok {
		utils.LogError("payload inválido para ReviveCardPayload")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*reviveCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	// revive a targetCard do targetPlayer
	err = targetPlayer.ReviveCard(*reviveCardPayload.TargetCardIndex)
	if err != nil {
		return err
	}

	return nil
}
