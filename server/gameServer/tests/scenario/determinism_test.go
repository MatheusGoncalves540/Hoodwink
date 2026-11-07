package scenario

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

// roomHash é o hash canônico do estado da sala (JSON). O relógio fixo (now) não é
// serializado, mas o efeito dele (StartTime, expiração, chat) aparece no JSON.
func roomHash(t *testing.T, r *rooms.Room) string {
	t.Helper()
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("serializar sala: %v", err)
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

// newSeededTable abre a mesa já com semente e relógio fixos (determinístico).
func newSeededTable(t *testing.T, seed rooms.Seed, now time.Time) *table {
	t.Helper()
	tb := newTable(t,
		seat("ana", 3, structs.CardAssassin, structs.CardKamikaze),
		seat("bia", 3, structs.CardGuardian, structs.CardInvestor),
		seat("cris", 3, structs.CardTrickster, structs.CardTrillionaire),
		seat("davi", 3, structs.CardPolitical, structs.CardRebel),
	)
	tb.room.Seed = seed
	tb.room.Rng = rooms.NewRngState(seed)
	tb.room.MatchID = "deterministic"
	tb.room.Created = now
	tb.room.StartTime = now
	if tb.room.GameEvent != nil {
		tb.room.GameEvent = structs.NewGameEvent(tb.room.GameEvent.PlayerID, tb.room.GameEvent.Type, now.Add(time.Minute), tb.room.GameEvent.Payload)
	}
	tb.room.SetNow(now)
	return tb
}

type step struct {
	who  string
	raw  string
	tick bool
}

// buildScript gera uma sequência determinística de jogadas (com jogadas inválidas no
// meio) e de ticks, a partir do índice da partida.
func buildScript(game int) []step {
	ids := []string{"ana", "bia", "cris", "davi"}
	rng := rand.New(rand.NewPCG(uint64(game)+1, 7))
	idx := func(r *rand.Rand) int { return r.IntN(3) } // 2 é índice inválido de propósito

	msgs := []func(*rand.Rand, string) string{
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
		func(_ *rand.Rand, to string) string {
			return fmt.Sprintf(`{"type":"TRICKSTER","payload":{"targetPlayer":%q}}`, to)
		},
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

	steps := make([]step, 0, 75)
	for range 50 {
		who := ids[rng.IntN(len(ids))]
		target := ids[rng.IntN(len(ids))]
		steps = append(steps, step{who: who, raw: msgs[rng.IntN(len(msgs))](rng, target)})
		if rng.IntN(3) == 0 {
			steps = append(steps, step{tick: true})
		}
	}
	return steps
}

// runGame joga o script contra uma sala com semente e relógio fixos, fazendo um
// Clone (ida/volta JSON) entre cada passo e devolvendo o hash de cada estado.
func runGame(t *testing.T, seed rooms.Seed, now time.Time, steps []step) []string {
	t.Helper()
	tb := newSeededTable(t, seed, now)

	hashes := make([]string, 0, len(steps)+1)
	hashes = append(hashes, roomHash(t, tb.room))

	for _, s := range steps {
		// Clone simula a ida/volta no Redis; o relógio fixo (now) não sobrevive ao clone.
		clone := tb.room.Clone()
		clone.SetNow(now)
		tb.room = clone

		before := roomHash(t, tb.room)

		if s.tick {
			engine.Tick(context.Background(), nil, registry, tb.room)
		} else {
			play, err := structs.ParsePlayerPlay([]byte(s.raw), s.who)
			if err == nil {
				if accepted := engine.ApplyPlay(context.Background(), nil, tb.room, play, registry); !accepted {
					// Jogada recusada não pode mudar o estado.
					if after := roomHash(t, tb.room); after != before {
						t.Fatalf("jogada recusada mudou o estado: %s", s.raw)
					}
				}
			}
		}

		hashes = append(hashes, roomHash(t, tb.room))
	}

	return hashes
}

// TestDeterministicReplaySameSeed joga 200 partidas duas vezes com a mesma semente,
// com Clone() entre passos, e exige estado idêntico (mesmo hash) a cada passo.
func TestDeterministicReplaySameSeed(t *testing.T) {
	now := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)

	for game := 0; game < 200; game++ {
		game := game
		t.Run(fmt.Sprint("partida ", game), func(t *testing.T) {
			var seed rooms.Seed
			binary.BigEndian.PutUint64(seed[:8], uint64(game))
			binary.BigEndian.PutUint64(seed[8:], 0xdeadbeef)

			steps := buildScript(game)
			run1 := runGame(t, seed, now, steps)
			run2 := runGame(t, seed, now, steps)

			if len(run1) != len(run2) {
				t.Fatalf("número de passos divergiu: %d vs %d", len(run1), len(run2))
			}
			for i := range run1 {
				if run1[i] != run2[i] {
					t.Fatalf("estado divergiu no passo %d", i)
				}
			}
		})
	}
}
