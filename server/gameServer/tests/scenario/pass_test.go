package scenario

import (
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

// A mesa de 3: um declara, os outros dois respondem em 1 s. Quando o último passa,
// a janela fecha na hora em vez de esperar o relógio de 10 s.
func TestPassFechaJanelaQuandoTodosRespondem(t *testing.T) {
	tb := newTable(t,
		seat("ana", 5, structs.CardTrillionaire, structs.CardRebel),
		seat("bia", 5, structs.CardAssassin, structs.CardRebel),
		seat("cris", 5, structs.CardGuardian, structs.CardRebel),
	)

	tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
	fimDaJanela := tb.room.GameEvent.ExpiresAt

	tb.mustPlay("bia", `{"type":"PASS"}`)
	if !tb.room.GameEvent.ExpiresAt.Equal(fimDaJanela) {
		t.Fatalf("um passe sozinho não pode fechar a janela")
	}

	tb.mustPlay("cris", `{"type":"PASS"}`)
	if tb.room.GameEvent.ExpiresAt.After(time.Now()) {
		t.Fatalf("com todos respondidos a janela deveria fechar na hora, vence em %s", tb.room.GameEvent.ExpiresAt)
	}

	// o processador resolve a janela fechada no próximo passo
	tb.tickUntil(structs.EventEarnCoins)
}

// Na carta morta só o dono da carta pode responder; é o passe dele que fecha a janela.
func TestPassNaCartaMortaSoOAlvoResponde(t *testing.T) {
	tb := newTable(t,
		seat("ana", 12, structs.CardAssassin, structs.CardRebel),
		seat("bia", 5, structs.CardTrickster, structs.CardRebel),
		seat("cris", 5, structs.CardTrillionaire, structs.CardRebel),
	)

	tb.mustPlay("ana", `{"type":"ASSASSIN","payload":{"targetPlayer":"cris","targetCardIndex":0}}`)
	tb.mustPlay("bia", `{"type":"PASS"}`)
	tb.mustPlay("cris", `{"type":"PASS"}`)
	tb.tickUntil(structs.EventCardKilled)

	// quem não é o alvo não tem reação: o passe dela não pode fechar a janela
	tb.mustPlay("bia", `{"type":"PASS"}`)
	if !tb.room.GameEvent.ExpiresAt.After(time.Now()) {
		t.Fatalf("passe de quem não é o alvo não pode fechar a janela da carta morta")
	}

	// o alvo responde e a janela fecha na hora
	tb.mustPlay("cris", `{"type":"PASS"}`)
	if tb.room.GameEvent.ExpiresAt.After(time.Now()) {
		t.Fatalf("alvo respondeu: a janela da carta morta deveria fechar")
	}
}

func TestPassRules(t *testing.T) {
	newT := func() *table {
		return newTable(t,
			seat("ana", 5, structs.CardTrillionaire, structs.CardRebel),
			seat("bia", 5, structs.CardAssassin, structs.CardRebel),
			seat("cris", 5, structs.CardGuardian, structs.CardRebel),
		)
	}

	t.Run("fora de janela de reação não há o que deixar passar", func(t *testing.T) {
		tb := newT()
		tb.mustReject("bia", `{"type":"PASS"}`)
	})

	t.Run("quem jogou não deixa passar a própria jogada", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustReject("ana", `{"type":"PASS"}`)
	})

	t.Run("morto não deixa passar", func(t *testing.T) {
		tb := newT()
		tb.p("cris").Alive = false
		tb.p("cris").Cards[0].Dead, tb.p("cris").Cards[1].Dead = true, true
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustReject("cris", `{"type":"PASS"}`)
	})

	t.Run("anúncio puro não é encerrado por passe", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TAKE_COINS"}`)
		tb.tick()
		if tb.eventType() != structs.EventEarnCoins {
			t.Fatalf("evento = %s, quer EARN_COINS", tb.eventType())
		}
		// a ação do turno não pode ser barrada; ninguém deveria estar respondendo
		// a este anúncio, e um passe solto não fecha a tela da mesa
		tb.mustPlay("bia", `{"type":"PASS"}`)
		if !tb.room.GameEvent.ExpiresAt.After(time.Now()) {
			t.Fatalf("um passe solto não pode encerrar o anúncio")
		}
	})

	t.Run("passes não vazam para a janela seguinte", func(t *testing.T) {
		tb := newT()
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustPlay("bia", `{"type":"PASS"}`)
		tb.mustPlay("cris", `{"type":"PASS"}`)
		tb.tickUntil(structs.EventEarnCoins) // abre o anúncio de moedas

		if tb.eventType() != structs.EventEarnCoins {
			tb.t.Fatalf("evento = %s, quer EARN_COINS", tb.eventType())
		}
		tb.mustPlay("bia", `{"type":"PASS"}`)
		if !tb.room.GameEvent.ExpiresAt.After(time.Now()) {
			t.Fatalf("o passe da janela anterior não pode fechar a janela nova")
		}
		tb.mustPlay("cris", `{"type":"PASS"}`)
		if tb.room.GameEvent.ExpiresAt.After(time.Now()) {
			t.Fatalf("respondido por todos, a janela nova também deveria fechar")
		}
	})

	t.Run("dois jogadores: o único adversário fecha a janela", func(t *testing.T) {
		tb := newTable(t,
			seat("ana", 5, structs.CardTrillionaire, structs.CardRebel),
			seat("bia", 5, structs.CardAssassin, structs.CardRebel),
		)
		tb.mustPlay("ana", `{"type":"TRILLIONAIRE"}`)
		tb.mustPlay("bia", `{"type":"PASS"}`)
		if tb.room.GameEvent.ExpiresAt.After(time.Now()) {
			t.Fatalf("com um só adversário o passe deveria fechar a janela")
		}
	})
}
