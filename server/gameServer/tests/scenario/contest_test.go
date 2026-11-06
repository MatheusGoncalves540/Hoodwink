package scenario

import (
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

// A captura do bug: Poppyin se diz Investidor e quem está na mesa duvida.
func TestContestInvestorBluffCancelsInvestmentAndCostsACard(t *testing.T) {
	tb := newTable(t,
		seat("pop", 7, structs.CardAssassin, structs.CardRebel),
		seat("por", 2, structs.CardKamikaze, structs.CardGravedigger),
	)

	tb.mustPlay("pop", `{"type":"INVESTOR"}`)
	tb.wantCoins("pop", 6) // a moeda investida sai da mão na declaração

	tb.mustPlay("por", `{"type":"CONTEST"}`)
	tb.tickUntil(structs.EventContestPenalty)

	// blefe: quem duvidou escolhe a carta do mentiroso
	if tb.room.GameEvent.PlayerID != "por" {
		t.Fatalf("quem escolhe = %s, quer por", tb.room.GameEvent.PlayerID)
	}
	tb.mustReject("pop", `{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":0}}`) // não é a vez dele de escolher
	tb.mustPlay("por", `{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":1}}`)
	tb.settle()

	tb.wantDead("pop", 1, true)
	if n := len(tb.p("pop").Investments); n != 0 {
		t.Fatalf("investimento de blefe foi aplicado (%d)", n)
	}
	tb.wantCoins("pop", 6) // a moeda do blefe não volta
}

func TestContestInvestorTruthKeepsInvestmentAndPunishesContestant(t *testing.T) {
	tb := newTable(t,
		seat("pop", 7, structs.CardInvestor, structs.CardRebel),
		seat("por", 2, structs.CardKamikaze, structs.CardGravedigger),
	)

	tb.mustPlay("pop", `{"type":"INVESTOR"}`)
	tb.mustPlay("por", `{"type":"CONTEST"}`)
	tb.tickUntil(structs.EventContestPenalty)

	if tb.room.GameEvent.PlayerID != "pop" {
		t.Fatalf("quem escolhe = %s, quer pop (tinha a carta)", tb.room.GameEvent.PlayerID)
	}
	tb.mustPlay("pop", `{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":1}}`)
	tb.settle()

	tb.wantDead("por", 1, true)
	tb.wantDead("pop", 0, false)
	if n := len(tb.p("pop").Investments); n != 1 {
		t.Fatalf("investimentos = %d, quer 1", n)
	}
}

func TestContestRules(t *testing.T) {
	newT := func() *table {
		return newTable(t,
			seat("ana", 5, structs.CardTrillionaire, structs.CardRebel),
			seat("bia", 5, structs.CardAssassin, structs.CardRebel),
			seat("cris", 5, structs.CardGuardian, structs.CardRebel),
		)
	}

	t.Run("ninguém contesta a própria carta", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustReject("ana", `{"type":"CONTEST"}`)
	})

	t.Run("sem carta declarada não há o que contestar", func(t *testing.T) {
		tb := newT()
		tb.mustReject("bia", `{"type":"CONTEST"}`)
	})

	t.Run("ação do turno não é contestável", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TAKE_COINS"}`)
		tb.mustReject("bia", `{"type":"CONTEST"}`)
	})

	t.Run("só uma contestação por declaração", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustPlay("bia", `{"type":"CONTEST"}`)
		tb.mustReject("cris", `{"type":"CONTEST"}`)
	})

	t.Run("morto não contesta", func(t *testing.T) {
		tb := newT()
		tb.p("cris").Alive = false
		tb.p("cris").Cards[0].Dead, tb.p("cris").Cards[1].Dead = true, true
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustReject("cris", `{"type":"CONTEST"}`)
	})

	t.Run("janela vencida não aceita contestação", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.tick() // a janela do Trilionário fechou
		tb.mustReject("bia", `{"type":"CONTEST"}`)
	})

	t.Run("verdade dita: o efeito acontece depois da penalidade", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustPlay("bia", `{"type":"CONTEST"}`)
		tb.tickUntil(structs.EventContestPenalty)
		tb.mustPlay("ana", `{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":0}}`)
		tb.settle()
		tb.wantDead("bia", 0, true)
		tb.wantCoins("ana", 8)
	})
}
