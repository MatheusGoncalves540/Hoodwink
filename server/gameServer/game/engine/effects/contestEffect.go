package effects

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ContestEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	// extrai o payload do efeito
	raw, ok := effect.Payload.(map[string]interface{})
	if !ok {
		utils.LogError("payload inválido")
		return
	}

	b, _ := json.Marshal(raw)

	var contestPayload structs.ContestPayload
	if err := json.Unmarshal(b, &contestPayload); err != nil {
		utils.LogError(err)
		return
	}

	// obtém o jogador que iniciou contestou
	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// obtém o jogador que está sendo contestado
	contestedPlayer, err := roomData.GetPlayer(contestPayload.ContestedPlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// Verifica se o jogador contestado possui a carta que está sendo contestada
	hasCard := false
	for _, card := range contestedPlayer.Cards {
		if string(card.Name) == contestPayload.ContestedCard {
			hasCard = true
			break
		}
	}

	// calcula o tempo de expiração do efeito
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	if err != nil {
		utils.LogError(err)
		return
	}

	// se o jogador contestado não possuir a carta, cancela o efeito pendente (o ultimo adicionado)
	if !hasCard {
		roomData.CancelLastPendingEffect(ctx, rdb)
	}

	penaltyPayload := structs.NewContestPenaltyPayload(contestedPlayer.Id, contestPayload.ContestedCard, hasCard, nil)

	// Jogador contestado/que contestou escolhe uma carta do outro jogador para matar (esse evento)
	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventContestPenalty, expiresAt, penaltyPayload)

	roomData.PendingEffects = append(roomData.PendingEffects,
		structs.Effect{
			Cause:        structs.EffectContestPenalty,
			SourcePlayer: sourcePlayer.Id,
			Payload:      penaltyPayload,
		},
	)

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
}
