package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func ContestEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	sink = rooms.SinkOrNoop(sink)

	contestPayload, ok := effect.Payload.(structs.ContestPayload)
	if !ok {
		pure.LogError("payload inválido para EffectContest")
		return
	}

	// jogador que contestou
	contestant, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		pure.LogError(err)
		return
	}

	// jogador que teve a carta contestada
	contestedPlayer, err := roomData.GetPlayer(contestPayload.ContestedPlayer)
	if err != nil {
		pure.LogError(err)
		return
	}

	// A carta só conta como "em mãos" se estiver viva.
	revealedCardIndex, hasCard := contestedPlayer.GetFirstAliveCardIndexByName(structs.CardName(contestPayload.ContestedCard))

	// Se o jogador não tinha a carta, o efeito da carta é cancelado.
	if !hasCard {
		roomData.PopLastPendingEffect()
	} else if err := contestedPlayer.RevealCardToAll(revealedCardIndex); err != nil {
		pure.LogError(err)
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
		pure.LogError(err)
		return
	}
	expiresAt := roomData.Now().Add(timeoutDuration * time.Second).UTC()

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
		pure.LogError(err)
		return
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
}
