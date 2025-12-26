package effects

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func KillCard(ctx context.Context, rdb *redis.Client, roomData *roomStructs.Room, effect roomStructs.Effect) {
	//verificação de payload
	var payload roomStructs.AssassinPayload
	payloadBytes, err := json.Marshal(effect.Payload)
	if err != nil {
		utils.LogError(err)
		return
	}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		utils.LogError(err)
		return
	}

	// verifica se o jogador alvo existe
	player, exists := roomData.Players[payload.TargetPlayer]
	if !exists {
		utils.LogError(fmt.Errorf("jogador alvo não encontrado: %s", payload.TargetPlayer))
		return
	}

	// verifica se o índice da carta é válido
	if payload.TargetCard < 0 || payload.TargetCard >= len(player.Cards) {
		utils.LogError(fmt.Errorf("índice de carta inválido: %d para jogador %s", payload.TargetCard, payload.TargetPlayer))
		return
	}

	// marca a targetCard do targetPlayer como morta (-1)
	roomData.Players[payload.TargetPlayer].Cards[payload.TargetCard] = -1

	expiresAt := time.Now().Add(7 * time.Second).UTC()
	roomData.PendingEvent = &roomStructs.PendingEvent{
		PlayerID:  effect.SourcePlayer,
		Type:      roomStructs.EventCardKilled,
		ExpiresAt: expiresAt, // TODO colocar tempo configuravel
		Payload: map[string]interface{}{
			"TargetPlayer": payload.TargetPlayer,
			"TargetCard":   payload.TargetCard,
			"Cause":        effect.Cause,
		},
	}
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
}
