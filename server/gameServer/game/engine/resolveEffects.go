package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func resolveNextEffect(ctx context.Context, rdb *redis.Client, roomData *roomStructs.Room) {
	effect := roomData.PendingEffects[0]

	switch effect.Cause {

	case roomStructs.EffectAssassin:
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
