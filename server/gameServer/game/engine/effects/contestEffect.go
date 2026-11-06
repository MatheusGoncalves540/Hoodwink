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

	// jogador que contestou
	contestant, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// jogador que teve a carta contestada
	contestedPlayer, err := roomData.GetPlayer(contestPayload.ContestedPlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// A carta só conta como "em mãos" se estiver viva.
	revealedCardIndex, hasCard := contestedPlayer.GetFirstAliveCardIndexByName(structs.CardName(contestPayload.ContestedCard))

	// Se o jogador não tinha a carta, o efeito da carta é cancelado.
	if !hasCard {
		roomData.PopLastPendingEffect()
	} else if err := contestedPlayer.RevealCardToAll(revealedCardIndex); err != nil {
		utils.LogError(err)
	}

	// Quem escolhe a carta que morre:
	// - se o jogador tinha a carta, ele mesmo escolhe uma carta do contestador;
	// - se não tinha, o contestador escolhe uma carta do jogador.
	chooserID := contestant.Id
	if hasCard {
		chooserID = contestedPlayer.Id
	}

	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction")
	if err != nil {
		utils.LogError(err)
		return
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	penaltyPayload := structs.NewContestPenaltyPayload(
		contestedPlayer.Id,
		contestant.Id,
		contestPayload.ContestedCard,
		hasCard,
		nil,
	)
	if hasCard {
		penaltyPayload.RevealedCardIndex = &revealedCardIndex
	}

	roomData.GameEvent = structs.NewGameEvent(chooserID, structs.EventContestPenalty, expiresAt, penaltyPayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectContestPenalty, chooserID, penaltyPayload)); err != nil {
		utils.LogError(err)
		return
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
}
