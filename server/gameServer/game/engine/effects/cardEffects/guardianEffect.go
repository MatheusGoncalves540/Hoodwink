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
	if protectedPlayer == nil {
		utils.LogError("protectedPlayer não encontrado para GuardianEffect")
		return
	}

	switch guardianPayload.ProtectedFrom {
	case "future":
		protectedCards := protectedPlayer.GetProtectedCardsIndexes()

		// se a carta já tiver sido protegida por Guardian, desprotege ela para que a nova carta seja protegida
		if len(protectedCards) > 0 {
			for _, protectedCardIndex := range protectedCards {
				protectedPlayer.UnprotectCard(protectedCardIndex)
			}
		}

		protectedPlayer.ProtectCard(guardianPayload.TargetCardIndex)

	case "death":
		roomData.PopLastPendingEffect()

	case "sight":
		roomData.PopLastPendingEffect()

	default:
		utils.LogError("guardianPayload.ProtectedFrom tem valor inválido: " + guardianPayload.ProtectedFrom)
		return
	}

}
