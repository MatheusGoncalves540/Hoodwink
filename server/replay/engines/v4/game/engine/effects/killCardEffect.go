package effects

import (
	"context"
	"slices"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

func KillCard(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	killCardPayload, ok := effect.Payload.(structs.KillCardPayload)
	if !ok {
		pure.LogError("payload inválido para KillCardPayload")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*killCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	if structs.ProtectionBlocks(structs.EffectCause(killCardPayload.Cause)) {
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
