package effects

import (
	"context"
	"slices"

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

	protectedCards := targetPlayer.GetProtectedCardsIndexes()
	if slices.Contains(protectedCards, *killCardPayload.TargetCardIndex) {
		// se a carta já estiver protegida, retira a proteção ao invés de matar a carta
		err = targetPlayer.UnprotectCard(*killCardPayload.TargetCardIndex)
		if err != nil {
			return err
		}

		return nil
	} else {
		// marca a targetCard do targetPlayer como morta
		err = targetPlayer.KillCard(*killCardPayload.TargetCardIndex)
		if err != nil {
			return err
		}

		// atualiza estado de morte do jogador caso a carta morta seja a última viva do jogador
		roomData.VerifyIfPlayerDead(targetPlayer)
	}

	return nil
}
