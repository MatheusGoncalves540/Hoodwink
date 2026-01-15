package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects/effectsValidations"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func KamikazeEffect(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	player, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// valida a jogada do jogador
	valid, err := effectsValidations.ValidateKamikaze(player)
	if err != nil || !valid {
		utils.LogInvldPlyrReq(err, player.Id)
		return
	}

	// resolve o efeito de matar carta
	err = KillCard(ctx, rdb, RegistryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}
