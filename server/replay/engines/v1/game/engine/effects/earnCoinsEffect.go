package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func EarnCoinsEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	earnCoinsPayload, ok := effect.Payload.(structs.EarnCoinsPayload)
	if !ok {
		pure.LogError("payload inválido para EffectEarnCoins")
		return
	}

	if earnCoinsPayload.EarnedCoins <= 0 {
		return
	}

	generalRules, err := roomData.GetGeneralRules(registryRules)
	if err != nil {
		pure.LogError(err)
		return
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		pure.LogError(err)
		return
	}

	// adiciona as moedas ao jogador e retorna breakLimit, que indica se o limite de moedas será ultrapassado
	breakLimit := sourcePlayer.AddCoins(earnCoinsPayload.EarnedCoins, *generalRules.MaxCoins)

	if breakLimit {
		// se o limite de moedas for ultrapassado, mata a carta viva de menor índice
		aliveIndexes := sourcePlayer.GetAliveCardsIndexes()
		if len(aliveIndexes) == 0 {
			return
		}

		greedPayload := structs.NewKillCardPayload(string(structs.EffectGreed), &sourcePlayer.Id, &aliveIndexes[0])

		killEffect := structs.Effect{
			Cause:        structs.EffectGreed,
			SourcePlayer: sourcePlayer.Id,
			Payload:      greedPayload,
		}

		KillAnnouncerCard(ctx, sink, registryRules, roomData, killEffect)
	}
}
