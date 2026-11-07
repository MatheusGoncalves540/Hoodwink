package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

func KamikazeEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	kamikazePayload, ok := effect.Payload.(structs.KamikazePayload)
	if !ok {
		pure.LogError("payload inválido para KamikazePayload")
		return
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		pure.LogError(err)
		return
	}

	// Se o jogador matou o próprio Kamikaze, a morte acontece somente na resolução do efeito,
	// depois da janela de contestação.
	if kamikazePayload.KilledHimSelf {
		selfKillIndex := findLiveKamikazeIndex(sourcePlayer, kamikazePayload.TargetAllyCardIndex)
		if selfKillIndex != nil {
			if err := sourcePlayer.KillCard(*selfKillIndex); err != nil {
				pure.LogError(err)
			} else {
				roomData.VerifyIfPlayerDead(sourcePlayer)
			}
		}
	}

	killPayload := structs.NewKillCardPayload(string(effect.Cause), kamikazePayload.TargetPlayer, kamikazePayload.TargetCardIndex)

	killEffect := structs.Effect{
		Cause:        structs.EffectKamikaze,
		SourcePlayer: effect.SourcePlayer,
		Payload:      killPayload,
	}

	if err := effects.KillAnnouncerCard(ctx, sink, registryRules, roomData, killEffect); err != nil {
		pure.LogError(err)
	}
}

// findLiveKamikazeIndex procura uma carta Kamikaze viva do jogador.
// O índice informado tem prioridade; se ele não for mais um Kamikaze (por exemplo,
// após uma contestação), procura outra cópia viva na mão.
func findLiveKamikazeIndex(player *players.Player, preferredIndex *int) *int {
	if preferredIndex != nil {
		if card, err := player.GetCardByIndex(*preferredIndex); err == nil && !card.Dead && card.Name == structs.CardKamikaze {
			return preferredIndex
		}
	}

	if index, ok := player.GetFirstAliveCardIndexByName(structs.CardKamikaze); ok {
		return &index
	}
	return nil
}
