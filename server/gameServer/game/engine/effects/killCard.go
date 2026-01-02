package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects/effectsValidations"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/mitchellh/mapstructure"
	"github.com/redis/go-redis/v9"
)

func KillCard(ctx context.Context, rdb *redis.Client, roomData *roomStructs.Room, effect roomStructs.Effect) {
	// decodifica o payload
	var payload roomStructs.KillCardPayload
	if err := mapstructure.Decode(effect.Payload, &payload); err != nil {
		utils.LogError(err) // Pegar todos esses logerror e tratar na função chamadora
		return
	}

	// valida o efeito
	valid, err := effectsValidations.ValidateKillCardEffect(roomData, effect, payload)
	if err != nil || !valid {
		utils.LogError(err)
		return
	}

	// marca a targetCard do targetPlayer como morta (-1)
	player, err := roomData.GetPlayer(*payload.TargetPlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	err = player.KillCard(*payload.TargetCardIndex)
	if err != nil {
		utils.LogError(err)
		return
	}

	// cria o evento pendente de carta morta
	expiresAt := time.Now().Add(7 * time.Second).UTC()
	roomData.GameEvent = &roomStructs.GameEvent{
		PlayerID:  effect.SourcePlayer,
		Type:      roomStructs.EventCardKilled,
		ExpiresAt: expiresAt, // TODO colocar tempo configuravel
		Payload: map[string]interface{}{
			"TargetPlayer": *payload.TargetPlayer,
			"TargetCard":   *payload.TargetCardIndex,
			"Cause":        effect.Cause,
		},
	}
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
}
