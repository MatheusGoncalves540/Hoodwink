package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects/effectsValidations"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func AssassinEffect(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomData *roomStructs.Room, effect roomStructs.Effect) {
	cardRules, err := roomData.GetCardRules(RegistryRules, string(effect.Cause))
	if err != nil {
		utils.LogError("Erro ao obter regras da carta Assassin: " + err.Error())
		return
	}

	player, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	valid, err := effectsValidations.ValidateAssassin(roomData, cardRules, effect, player)
	if err != nil || !valid {
		utils.LogInvldPlyrReq(err)
		return
	}

	// remove coins do jogador
	player.RemoveCoins(*cardRules.Price)

	// resolve o efeito de matar carta
	err = KillCard(ctx, rdb, RegistryRules, roomData, effect)
	if err != nil {
		utils.LogError(err)
		return
	}
}
