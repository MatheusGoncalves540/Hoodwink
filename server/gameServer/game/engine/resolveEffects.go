package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func resolveNextEffect(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomData *roomStructs.Room) {
	effect := roomData.PendingEffects[0]

	switch effect.Cause {

	case roomStructs.EffectAssassin:
		cardRules, err := roomData.GetCardRules(RegistryRules, string(effect.Cause))
		if err != nil {
			utils.LogError("Erro ao obter regras da carta Assassin: " + err.Error())
			return
		}

		player := roomData.Players[effect.SourcePlayer]
		player.Coins -= *cardRules.Price
		roomData.Players[effect.SourcePlayer] = player

		effects.KillCard(ctx, rdb, roomData, effect)

		// TODO se matou → adiciona efeito de kamikaze
		// roomData.PendingEffects = append(roomData.PendingEffects,
		// 	roomStructs.Effect{
		// 		Type:         roomStructs.EffectKamikaze,
		// 		TargetPlayer: effect.TargetPlayer,
		// 	},
		// )

		// case roomStructs.EffectKamikaze:
		// 	// abre janela de decisão
		// 	roomData.State = roomStructs.StateWaitKamikaze
		// 	return
	}
	// remove o efeito que está sendo resolvido
	roomData.PendingEffects = roomData.PendingEffects[1:]
}
