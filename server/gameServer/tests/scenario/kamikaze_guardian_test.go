package scenario

import (
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

func TestKamikazeOnOwnTurnSacrificesItselfAndKills(t *testing.T) {
	tb := newTable(t,
		seat("ana", 0, structs.CardKamikaze, structs.CardRebel),
		seat("bia", 3, structs.CardAssassin, structs.CardRebel),
	)

	// sem dizer qual Kamikaze morre o servidor recusa (era o que o cliente mandava)
	tb.mustReject("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0}}`)
	// carta própria que não é Kamikaze não pode ser sacrificada
	tb.mustReject("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0,"targetAllyCardIndex":1}}`)
	// não mira em si mesmo
	tb.mustReject("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"ana","targetCardIndex":1,"targetAllyCardIndex":0}}`)

	tb.mustPlay("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0,"targetAllyCardIndex":0}}`)
	tb.settle()

	tb.wantDead("ana", 0, true)
	tb.wantDead("bia", 0, true)
	tb.wantCoins("ana", 0) // de graça
}

func TestKamikazeRetaliatesWhenKilled(t *testing.T) {
	tb := newTable(t,
		seat("bia", 3, structs.CardAssassin, structs.CardRebel),
		seat("ana", 0, structs.CardKamikaze, structs.CardRebel),
	)

	tb.mustPlay("bia", `{"type":"ASSASSIN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`)
	tb.tickUntil(structs.EventCardKilled)

	tb.mustPlay("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":1}}`)
	tb.settle()

	tb.wantDead("ana", 0, true)
	tb.wantDead("bia", 1, true)
}

func TestKamikazeReactionRules(t *testing.T) {
	t.Run("só a carta Kamikaze que está morrendo reage", func(t *testing.T) {
		tb := newTable(t,
			seat("bia", 3, structs.CardAssassin, structs.CardRebel),
			seat("ana", 0, structs.CardKamikaze, structs.CardRebel),
		)
		tb.mustPlay("bia", `{"type":"ASSASSIN","payload":{"targetPlayer":"ana","targetCardIndex":1}}`)
		tb.tickUntil(structs.EventCardKilled)
		tb.mustReject("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0}}`)
	})

	t.Run("quem não é o alvo não reage", func(t *testing.T) {
		tb := newTable(t,
			seat("bia", 3, structs.CardAssassin, structs.CardRebel),
			seat("ana", 0, structs.CardRebel, structs.CardRebel),
			seat("cris", 0, structs.CardKamikaze, structs.CardRebel),
		)
		tb.mustPlay("bia", `{"type":"ASSASSIN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`)
		tb.tickUntil(structs.EventCardKilled)
		tb.mustReject("cris", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0}}`)
	})

	t.Run("Guardiã não segura o golpe do Kamikaze em cima da hora", func(t *testing.T) {
		tb := newTable(t,
			seat("ana", 9, structs.CardKamikaze, structs.CardRebel),
			seat("bia", 9, structs.CardGuardian, structs.CardRebel),
		)
		tb.mustPlay("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0,"targetAllyCardIndex":0}}`)
		tb.tickUntil(structs.EventCardKilled)
		tb.mustReject("bia", `{"type":"GUARDIAN","payload":{"targetPlayer":"bia","targetCardIndex":0}}`)
		tb.settle()
		tb.wantDead("bia", 0, true)
	})

	// rules.html: "Caso a carta já estivesse protegida antes, o ataque pode ser parado."
	t.Run("proteção feita antes para o Kamikaze", func(t *testing.T) {
		tb := newTable(t,
			seat("bia", 9, structs.CardGuardian, structs.CardRebel),
			seat("ana", 0, structs.CardKamikaze, structs.CardRebel),
		)
		tb.mustPlay("bia", `{"type":"GUARDIAN","payload":{"targetPlayer":"bia","targetCardIndex":0}}`)
		tb.settle()
		tb.mustPlay("ana", `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0,"targetAllyCardIndex":0}}`)
		tb.settle()
		tb.wantDead("ana", 0, true)
		tb.wantDead("bia", 0, false)
		if tb.p("bia").Cards[0].Protected {
			t.Fatal("a proteção deveria ter sido gasta")
		}
	})
}

func TestGuardianProtectsForReal(t *testing.T) {
	t.Run("proteção do turno barra o próximo ataque e é gasta", func(t *testing.T) {
		tb := newTable(t,
			seat("ana", 9, structs.CardGuardian, structs.CardRebel),
			seat("bia", 9, structs.CardAssassin, structs.CardRebel),
		)
		tb.mustPlay("ana", `{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":1}}`)
		tb.settle()
		tb.wantCoins("ana", 9-6) // no turno: 5 + 1 de sobretaxa, taxa 0
		if !tb.p("ana").Cards[1].Protected {
			t.Fatal("carta não ficou protegida")
		}

		tb.mustPlay("bia", `{"type":"ASSASSIN","payload":{"targetPlayer":"ana","targetCardIndex":1}}`)
		tb.settle()
		tb.wantDead("ana", 1, false)
		if tb.p("ana").Cards[1].Protected {
			t.Fatal("a proteção deveria ter sido gasta")
		}
	})

	t.Run("reação na hora do ataque salva a carta", func(t *testing.T) {
		tb := newTable(t,
			seat("bia", 9, structs.CardAssassin, structs.CardRebel),
			seat("ana", 9, structs.CardGuardian, structs.CardRebel),
		)
		tb.mustPlay("bia", `{"type":"ASSASSIN","payload":{"targetPlayer":"ana","targetCardIndex":1}}`)
		tb.tickUntil(structs.EventCardKilled)
		tb.mustReject("ana", `{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`) // carta errada
		tb.mustPlay("ana", `{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":1}}`)
		tb.settle()
		tb.wantDead("ana", 1, false)
		tb.wantCoins("ana", 9-5)
	})

	t.Run("contestação não é bloqueável", func(t *testing.T) {
		tb := newTable(t,
			seat("bia", 9, structs.CardTrillionaire, structs.CardRebel),
			seat("ana", 9, structs.CardGuardian, structs.CardRebel),
		)
		tb.p("ana").Cards[0].Protected = true
		tb.mustPlay("bia", `{"type":"TRILLIONAIRE"}`)
		tb.mustPlay("ana", `{"type":"CONTEST"}`)
		tb.tickUntil(structs.EventContestPenalty)
		tb.mustPlay("bia", `{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":0}}`)
		tb.settle()
		tb.wantDead("ana", 0, true)
	})

	t.Run("Guardiã bloqueia o Clarividente", func(t *testing.T) {
		tb := newTable(t,
			seat("bia", 9, structs.CardClairvoyant, structs.CardRebel),
			seat("ana", 9, structs.CardGuardian, structs.CardRebel),
		)
		tb.mustPlay("bia", `{"type":"CLAIRVOYANT","payload":{"targetPlayer":"ana","targetCardIndex":1,"showToAllPlayers":true}}`)
		tb.mustPlay("ana", `{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":1}}`)
		tb.settle()
		if tb.p("ana").Cards[1].RevealedToAll {
			t.Fatal("carta foi revelada apesar da Guardiã")
		}
		tb.wantCoins("ana", 9-4) // contra Clarividente: 5 − 1, taxa 0
	})
}
