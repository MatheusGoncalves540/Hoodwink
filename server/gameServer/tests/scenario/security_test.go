package scenario

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

func twoSeats(t testing.TB) *table {
	return newTable(t,
		seat("ana", 8, structs.CardAssassin, structs.CardGuardian),
		seat("bia", 8, structs.CardKamikaze, structs.CardTrickster),
	)
}

func TestRejectsMalformedOrForgedMessages(t *testing.T) {
	bad := map[string]string{
		"JSON quebrado":             `{"type":"ASSASSIN",`,
		"campo a mais":              `{"type":"TAKE_COINS","payload":{"coins":20}}`,
		"tipo desconhecido":         `{"type":"GIVE_ME_COINS"}`,
		"sem tipo":                  `{"payload":{}}`,
		"índice negativo":           `{"type":"ASSASSIN","payload":{"targetPlayer":"bia","targetCardIndex":-1}}`,
		"índice fora da mão":        `{"type":"ASSASSIN","payload":{"targetPlayer":"bia","targetCardIndex":99}}`,
		"sem alvo":                  `{"type":"ASSASSIN","payload":{"targetCardIndex":0}}`,
		"alvo inexistente":          `{"type":"ASSASSIN","payload":{"targetPlayer":"zé","targetCardIndex":0}}`,
		"Guardiã na carta de outro": `{"type":"GUARDIAN","payload":{"targetPlayer":"bia","targetCardIndex":0}}`,
		"Trapaceiro em si mesmo":    `{"type":"TRICKSTER","payload":{"targetPlayer":"ana"}}`,
		"payload no lugar errado":   `{"type":"KAMIKAZE","payload":"bia"}`,
	}
	for name, raw := range bad {
		t.Run(name, func(t *testing.T) {
			tb := twoSeats(t)
			tb.mustReject("ana", raw)
			tb.wantCoins("ana", 8)
			tb.wantDead("bia", 0, false)
		})
	}
}

func TestPlayerIdComesFromTheConnectionNotTheMessage(t *testing.T) {
	tb := twoSeats(t)
	// ana tenta se passar por bia: o servidor usa o id do ticket e ignora o do corpo
	tb.mustPlay("ana", `{"type":"TAKE_COINS","playerId":"bia"}`)
	tb.settle()
	tb.wantCoins("ana", 9)
	tb.wantCoins("bia", 8)
}

func TestTurnOwnership(t *testing.T) {
	tb := twoSeats(t)
	tb.mustReject("bia", `{"type":"TAKE_COINS"}`)
	tb.mustReject("bia", `{"type":"ASSASSIN","payload":{"targetPlayer":"ana","targetCardIndex":0}}`)
	tb.mustPlay("ana", `{"type":"TAKE_COINS"}`)
	tb.mustReject("ana", `{"type":"TAKE_COINS"}`) // uma ação por turno
	tb.settle()
	tb.wantCoins("ana", 9)
	if tb.room.CurrentPlayer != "bia" {
		t.Fatalf("vez de %s, quer bia", tb.room.CurrentPlayer)
	}
}

func TestNoPlaysAfterGameOver(t *testing.T) {
	tb := twoSeats(t)
	tb.room.GameOver = true
	tb.mustReject("ana", `{"type":"TAKE_COINS"}`)
}

// Partida aleatória longa: nenhuma mensagem pode derrubar o motor nem quebrar os invariantes.
func TestSoakRandomGameKeepsInvariants(t *testing.T) {
	ids := []string{"ana", "bia", "cris", "davi"}
	type gen func(r *rand.Rand, target string) string
	idx := func(r *rand.Rand) int { return r.IntN(3) } // 2 é índice inválido de propósito
	msgs := []gen{
		func(*rand.Rand, string) string { return `{"type":"TAKE_COINS"}` },
		func(*rand.Rand, string) string { return `{"type":"TRILLIONAIRE"}` },
		func(*rand.Rand, string) string { return `{"type":"INVESTOR"}` },
		func(*rand.Rand, string) string { return `{"type":"POLITICAL"}` },
		func(*rand.Rand, string) string { return `{"type":"REBEL"}` },
		func(*rand.Rand, string) string { return `{"type":"CONTEST"}` },
		func(*rand.Rand, string) string { return `{"type":"SELFISH"}` },
		func(r *rand.Rand, _ string) string {
			return fmt.Sprintf(`{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":%d}}`, idx(r))
		},
		func(r *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"ASSASSIN","payload":{"targetPlayer":%q,"targetCardIndex":%d}}`, to, idx(r))
		},
		func(r *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"KILL_CARD","payload":{"targetPlayer":%q,"targetCardIndex":%d}}`, to, idx(r))
		},
		func(r *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"KAMIKAZE","payload":{"targetPlayer":%q,"targetCardIndex":%d,"targetAllyCardIndex":%d}}`, to, idx(r), idx(r))
		},
		func(r *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"CLAIRVOYANT","payload":{"targetPlayer":%q,"targetCardIndex":%d,"showToAllPlayers":true}}`, to, idx(r))
		},
		func(_ *rand.Rand, to string) string { return fmt.Sprintf(`{"type":"TRICKSTER","payload":{"targetPlayer":%q}}`, to) },
		func(r *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"GUARDIAN","payload":{"targetPlayer":%q,"targetCardIndex":%d}}`, to, idx(r))
		},
		func(r *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"GRAVEDIGGER","payload":{"targetPlayer":%q,"targetCardIndex":%d}}`, to, idx(r))
		},
		func(r *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"CROUPIER","payload":{"targetPlayer":%q,"targetCardIndex":%d,"useOnTwoCards":false}}`, to, idx(r))
		},
		func(r *rand.Rand, _ string) string {
			return fmt.Sprintf(`{"type":"SWAP_CARDS","payload":{"targetCardIndex":%d}}`, idx(r))
		},
	}

	all := []structs.CardName{structs.CardAssassin, structs.CardKamikaze, structs.CardGuardian, structs.CardInvestor, structs.CardTrickster, structs.CardTrillionaire}
	accepted := 0
	for seed := range uint64(20) {
		t.Run(fmt.Sprint("semente ", seed), func(t *testing.T) {
			rng := rand.New(rand.NewPCG(seed, 7))
			card := func() structs.CardName { return all[rng.IntN(len(all))] }
			tb := newTable(t,
				seat(ids[0], 3, card(), card()), seat(ids[1], 3, card(), card()),
				seat(ids[2], 3, card(), card()), seat(ids[3], 3, card(), card()),
			)
			for range 400 {
				if tb.room.GameOver {
					break
				}
				who, target := ids[rng.IntN(len(ids))], ids[rng.IntN(len(ids))]
				if tb.play(who, msgs[rng.IntN(len(msgs))](rng, target)) {
					accepted++
				}
				if rng.IntN(3) == 0 {
					tb.tick()
				}
				checkInvariants(t, tb)
			}
		})
	}
	// sem jogadas aceitas o teste não exercitaria nada
	if accepted < 200 {
		t.Fatalf("só %d jogadas aceitas em 20 partidas", accepted)
	}
}

func checkInvariants(t *testing.T, tb *table) {
	t.Helper()
	for id, p := range tb.room.Players {
		if p.Coins < 0 {
			t.Fatalf("%s ficou com %d moedas", id, p.Coins)
		}
		alive := false
		for _, c := range p.Cards {
			alive = alive || !c.Dead
			if c.Dead && c.Protected {
				t.Fatalf("%s tem carta morta protegida", id)
			}
		}
		if p.Alive != alive {
			t.Fatalf("%s: Alive=%v mas tem carta viva=%v", id, p.Alive, alive)
		}
	}
}

// go test -fuzz=FuzzPlay ./tests/scenario (os casos-semente rodam em todo go test)
func FuzzPlay(f *testing.F) {
	for _, s := range []string{
		`{"type":"INVESTOR"}`, `{"type":"CONTEST"}`, `{"type":"KAMIKAZE","payload":{"targetPlayer":"bia","targetCardIndex":0,"targetAllyCardIndex":0}}`,
		`{"type":"GUARDIAN","payload":{"targetPlayer":"ana","targetCardIndex":1}}`, `{"type":"CONTEST_PENALTY","payload":{"TargetCardIndex":0}}`,
		`{"type":"COMMAND_START_GAME","payload":{}}`, `{}`, `[]`, `null`,
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		tb := twoSeats(t)
		tb.play("ana", raw)
		tb.tick()
		tb.play("bia", raw)
		checkInvariants(t, tb)
	})
}
