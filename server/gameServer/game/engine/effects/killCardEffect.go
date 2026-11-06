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

// protectionCanBlockKill indica se a proteção da Guardiã pode impedir a morte
// de acordo com a causa do ataque. Kamikaze, contestação e perda de carta por
// limite de moedas não podem ser impedidos.
func protectionCanBlockKill(cause string) bool {
	switch structs.EffectCause(cause) {
	case structs.EffectKamikaze, structs.EffectContestPenalty, structs.EffectGreed:
		return false
	default:
		return true
	}
}

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

	if protectionCanBlockKill(killCardPayload.Cause) {
		protectedCards := targetPlayer.GetProtectedCardsIndexes()
		if slices.Contains(protectedCards, *killCardPayload.TargetCardIndex) {
			// se a carta estava protegida, a Guardiã consome a proteção ao invés de matar a carta
			return targetPlayer.UnprotectCard(*killCardPayload.TargetCardIndex)
		}
	}

	// marca a targetCard do targetPlayer como morta
	if err := targetPlayer.KillCard(*killCardPayload.TargetCardIndex); err != nil {
		return err
	}

	// atualiza estado de morte do jogador caso a carta morta seja a última viva do jogador
	roomData.VerifyIfPlayerDead(targetPlayer)

	return nil
}
