package cardEffects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func GravediggerEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	gravediggerPayload, ok := effect.Payload.(structs.GravediggerPayload)
	if !ok {
		utils.LogError("payload inválido para GravediggerPayload")
		return
	}

	changeCardPayload := structs.NewChangeCardPayload(string(effect.Cause), gravediggerPayload.TargetPlayer, gravediggerPayload.TargetCardIndex, nil)

	changeCardEffect := structs.Effect{
		Cause:        structs.EffectChangedCard,
		SourcePlayer: effect.SourcePlayer,
		Payload:      changeCardPayload,
	}

	// chama o ChangeCardAnnouncer para trocar a carta do jogador alvo
	err := effects.ChangeCardAnnouncer(ctx, rdb, registryRules, roomData, changeCardEffect)
	if err != nil {
		utils.LogError(err)
		return
	}
}
