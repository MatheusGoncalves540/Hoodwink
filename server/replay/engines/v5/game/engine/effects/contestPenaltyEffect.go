package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

// ContestPenaltyEffect aplica a penalidade da contestação.
// Se o jogador contestado tinha a carta, ele escolhe uma carta do contestador para morrer
// e depois descarta a carta revelada, comprando outra do baralho.
// Se não tinha, o contestador escolhe uma carta do jogador contestado para morrer.
func ContestPenaltyEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	contestPenaltyPayload, ok := effect.Payload.(structs.ContestPenaltyPayload)
	if !ok {
		pure.LogError("payload inválido para EffectContestPenalty")
		return
	}

	contestedPlayer, err := roomData.GetPlayer(contestPenaltyPayload.ContestedPlayer)
	if err != nil {
		pure.LogError(err)
		return
	}

	contestingPlayer, err := roomData.GetPlayer(contestPenaltyPayload.ContestingPlayer)
	if err != nil {
		pure.LogError(err)
		return
	}

	// Define quem perde a carta: o contestador, se o jogador tinha a carta;
	// caso contrário, o próprio jogador contestado.
	targetPlayer := contestedPlayer
	if contestPenaltyPayload.HasCard {
		targetPlayer = contestingPlayer
	}

	targetIndex := pickAliveCardIndex(roomData, targetPlayer, contestPenaltyPayload.TargetCardIndex)
	if targetIndex == nil {
		pure.LogError("nenhuma carta viva disponível para a penalidade de contestação")
		return
	}

	killPayload := structs.NewKillCardPayload(string(structs.EffectContestPenalty), &targetPlayer.Id, targetIndex)

	killEffect := structs.NewEffect(structs.EffectContestPenalty, effect.SourcePlayer, killPayload)
	if err := KillAnnouncerCard(ctx, sink, registryRules, roomData, killEffect); err != nil {
		pure.LogError(err)
		return
	}

	// Se o jogador realmente tinha a carta, ela é revelada, descartada e substituída
	// por outra carta do baralho, sem mudar de posição.
	if !contestPenaltyPayload.HasCard {
		return
	}

	revealedIndex, ok := resolveRevealedCardIndex(contestedPlayer, contestPenaltyPayload)
	if !ok {
		return
	}

	changeCardPayload := structs.NewChangeCardPayload(string(structs.EffectContestPenalty), &contestedPlayer.Id, &revealedIndex, nil)

	changeCardEffect := structs.NewEffect(structs.EffectContestPenalty, effect.SourcePlayer, changeCardPayload)
	if err := ChangeCardAnnouncer(ctx, sink, registryRules, roomData, changeCardEffect); err != nil {
		pure.LogError(err)
	}
}

// resolveRevealedCardIndex usa o índice da carta revelada calculado pelo servidor.
// Se ele não estiver mais disponível, procura a primeira carta viva com o nome contestado.
func resolveRevealedCardIndex(player *players.Player, payload structs.ContestPenaltyPayload) (int, bool) {
	if payload.RevealedCardIndex != nil {
		card, err := player.GetCardByIndex(*payload.RevealedCardIndex)
		if err == nil && !card.Dead {
			return *payload.RevealedCardIndex, true
		}
	}

	return player.GetFirstAliveCardIndexByName(structs.CardName(payload.ContestedCard))
}

// pickAliveCardIndex valida o índice escolhido e, caso seja inválido ou a carta
// esteja morta, escolhe aleatoriamente uma carta viva do jogador alvo.
func pickAliveCardIndex(roomData *rooms.Room, player *players.Player, chosen *int) *int {
	if chosen != nil {
		card, err := player.GetCardByIndex(*chosen)
		if err == nil && !card.Dead {
			return chosen
		}
	}

	aliveIndexes := player.GetAliveCardsIndexes()
	if len(aliveIndexes) == 0 {
		return nil
	}
	index := aliveIndexes[roomData.Rand().IntN(len(aliveIndexes))]
	return &index
}
