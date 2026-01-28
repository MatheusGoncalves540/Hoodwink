package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func KillCard(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	killCardPayload, ok := effect.Payload.(structs.KillCardPayload)
	if !ok {
		utils.LogError("payload inválido para KillCardPayload")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*killCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	// marca a targetCard do targetPlayer como morta
	err = targetPlayer.KillCard(*killCardPayload.TargetCardIndex)
	if err != nil {
		return err
	}

	return nil
}
