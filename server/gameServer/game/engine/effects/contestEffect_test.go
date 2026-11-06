package effects

import (
	"context"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func loadClassicRulesForEffectsTest(t *testing.T) *rules.Registry {
	t.Helper()

	gameRules, err := rules.LoadRules("../../rules/modes/ClassicRules.yaml")
	if err != nil {
		t.Fatalf("erro ao carregar ClassicRules.yaml: %v", err)
	}
	registry := rules.NewRegistry()
	registry.Register(string(structs.ClassicRules), gameRules)
	return registry
}

func newContestRoom() *rooms.Room {
	return &rooms.Room{
		Rules:       structs.ClassicRules,
		TimeoutType: string(rules.TimeoutsTypeDefault),
		Players: map[string]*players.Player{
			"p1": {
				Id:    "p1",
				Cards: []structs.Card{{Name: structs.CardAssassin, Index: 0}},
				Alive: true,
			},
			"p2": {
				Id:    "p2",
				Cards: []structs.Card{{Name: structs.CardRebel, Index: 0}},
				Alive: true,
			},
		},
		Deck: []structs.CardName{structs.CardGuardian, structs.CardTrickster},
	}
}

func TestContestEffectWhenPlayerHadCard(t *testing.T) {
	registry := loadClassicRulesForEffectsTest(t)
	room := newContestRoom()
	room.PendingEffects = []structs.EffectDTO{{Cause: structs.EffectAssassin, SourcePlayer: "p1"}}

	effect := structs.NewEffect(structs.EffectContest, "p2", structs.NewContestPayload("p1", string(structs.CardAssassin)))
	ContestEffect(context.Background(), nil, registry, room, effect)

	if room.GameEvent == nil || room.GameEvent.Type != structs.EventContestPenalty {
		t.Fatal("evento de penalidade de contestação não foi aberto")
	}
	if room.GameEvent.PlayerID != "p1" {
		t.Fatalf("quem deve escolher = %q, quer p1 (jogador que tinha a carta)", room.GameEvent.PlayerID)
	}
	if len(room.PendingEffects) != 2 {
		t.Fatalf("efeitos pendentes = %d, quer 2 (efeito da carta + penalidade)", len(room.PendingEffects))
	}
	if !room.Players["p1"].Cards[0].RevealedToAll {
		t.Fatal("a carta contestada que o jogador possuía deveria ser revelada para todos")
	}
}

func TestContestEffectWhenPlayerDidNotHaveCardCancelsEffect(t *testing.T) {
	registry := loadClassicRulesForEffectsTest(t)
	room := newContestRoom()
	room.Players["p1"].Cards[0].Name = structs.CardTrickster
	room.PendingEffects = []structs.EffectDTO{{Cause: structs.EffectAssassin, SourcePlayer: "p1"}}

	effect := structs.NewEffect(structs.EffectContest, "p2", structs.NewContestPayload("p1", string(structs.CardAssassin)))
	ContestEffect(context.Background(), nil, registry, room, effect)

	if room.GameEvent == nil || room.GameEvent.PlayerID != "p2" {
		t.Fatal("o contestador deveria ser quem escolhe a carta quando o jogador não tinha a carta")
	}
	if len(room.PendingEffects) != 1 {
		t.Fatalf("efeitos pendentes = %d, quer 1 (apenas a penalidade)", len(room.PendingEffects))
	}
}

func TestContestPenaltyEffectKillsChosenCardAndReplacesRevealedCard(t *testing.T) {
	registry := loadClassicRulesForEffectsTest(t)
	room := newContestRoom()

	target := 0
	payload := structs.NewContestPenaltyPayload("p1", "p2", string(structs.CardAssassin), true, &target)
	effect := structs.NewEffect(structs.EffectContestPenalty, "p1", payload)

	ContestPenaltyEffect(context.Background(), nil, registry, room, effect)

	// Duas apresentações: morte da carta do contestador e troca da carta revelada.
	if len(room.PendingPresentationEvents) != 2 {
		t.Fatalf("eventos de apresentação = %d, quer 2", len(room.PendingPresentationEvents))
	}
	if room.PendingPresentationEvents[0].Type != structs.EventCardKilled {
		t.Fatalf("primeiro evento = %s, quer %s", room.PendingPresentationEvents[0].Type, structs.EventCardKilled)
	}
	if room.PendingPresentationEvents[1].Type != structs.EventChangedCard {
		t.Fatalf("segundo evento = %s, quer %s", room.PendingPresentationEvents[1].Type, structs.EventChangedCard)
	}
}
