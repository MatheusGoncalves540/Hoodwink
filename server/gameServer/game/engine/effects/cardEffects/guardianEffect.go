package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func GuardianEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	// resolve o efeito de proteger carta
	guardianPayload, ok := effect.Payload.(structs.GuardianPayload)
	if !ok {
		utils.LogError("payload inválido para GuardianPayload")
		return
	}

	protectedPlayer, err := roomData.GetPlayer(guardianPayload.TargetPlayer)
	if err != nil {
		utils.LogError("protectedPlayer não encontrado para GuardianEffect: " + err.Error())
		return
	}

	switch guardianPayload.ProtectedFrom {
	case "future":
		protectedCards := protectedPlayer.GetProtectedCardsIndexes()

		// se a carta já tiver sido protegida por Guardian, desprotege para que a nova carta seja protegida
		for _, protectedCardIndex := range protectedCards {
			if err := protectedPlayer.UnprotectCard(protectedCardIndex); err != nil {
				utils.LogError(err)
			}
		}

		if err := protectedPlayer.ProtectCard(guardianPayload.TargetCardIndex); err != nil {
			utils.LogError(err)
		}

	case "death":
		if cancelLastPendingEffect(roomData, structs.EffectCardKilled) {
			return
		}
		utils.LogError("Guardian tentou bloquear uma morte, mas o efeito não estava pendente")

	case "sight":
		if cancelLastPendingEffect(roomData, structs.EffectClairvoyant) {
			return
		}
		utils.LogError("Guardian tentou bloquear um Clairvoyant, mas o efeito não estava pendente")

	default:
		utils.LogError("guardianPayload.ProtectedFrom tem valor inválido: " + guardianPayload.ProtectedFrom)
		return
	}
}

// cancelLastPendingEffect remove o último efeito pendente apenas se ele for do tipo esperado.
func cancelLastPendingEffect(roomData *rooms.Room, expected structs.EffectCause) bool {
	last, ok := roomData.PeekLastPendingEffect()
	if !ok || last.Cause != expected {
		return false
	}
	roomData.PopLastPendingEffect()
	return true
}
