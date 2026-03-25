package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ContestEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	contestPayload, ok := effect.Payload.(structs.ContestPayload)
	if !ok {
		utils.LogError("payload inválido para EffectContest")
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
		roomData.PopLastPendingEffect()
	}

	penaltyPayload := structs.NewContestPenaltyPayload(contestedPlayer.Id, contestPayload.ContestedCard, hasCard, nil)

	// Jogador contestado/que contestou escolhe uma carta do outro jogador para matar (esse evento)
	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventContestPenalty, expiresAt, penaltyPayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectContestPenalty, sourcePlayer.Id, penaltyPayload)); err != nil {
		utils.LogError(err)
		return
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
}
