// Package scenario roda partidas inteiras do motor em memória, sem Redis:
// JSON do cliente → ParsePlayerPlay → validação → protocolo → engine.Tick.
//
// Cada jogada e cada tick passam a sala por JSON, como o servidor faz ao ler/gravar no Redis
// (as validações leem GameEvent.Payload como map, que só existe depois desse caminho).
//
//	tb := newTable(t, seat("ana", 3, structs.CardInvestor, structs.CardAssassin), seat("bia", 3, ...))
//	tb.mustPlay("ana", `{"type":"INVESTOR"}`)
//	tb.mustPlay("bia", `{"type":"CONTEST"}`)
//	tb.tickUntil(structs.EventContestPenalty)
package scenario

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

var registry = func() *rules.Registry {
	gameRules, err := rules.LoadRules("../../game/rules/modes/ClassicRules.yaml")
	if err != nil {
		panic(err)
	}
	reg := rules.NewRegistry()
	reg.Register(string(structs.ClassicRules), gameRules)
	return reg
}()

type seatSpec struct {
	id    string
	coins int
	cards []structs.CardName
}

func seat(id string, coins int, cards ...structs.CardName) seatSpec {
	return seatSpec{id: id, coins: coins, cards: cards}
}

type table struct {
	t    testing.TB
	room *rooms.Room
}

// newTable abre a partida na vez do primeiro assento, com taxa 0 e um baralho neutro.
func newTable(t testing.TB, seats ...seatSpec) *table {
	t.Helper()
	room := &rooms.Room{
		ID:                "SCENARIO",
		Rules:             structs.ClassicRules,
		TimeoutType:       string(rules.TimeoutsTypeDefault),
		MaxPlayers:        10,
		Turn:              1,
		Players:           map[string]*players.Player{},
		DoubledCardValues: map[structs.TypePlayerPlays]structs.DoubledCardValues{},
		Deck:              []structs.CardName{structs.CardRebel, structs.CardPolitical, structs.CardTrickster, structs.CardCroupier},
		CurrentPlayer:     seats[0].id,
		StartTime:         time.Now().UTC(),
	}
	for i, s := range seats {
		p := &players.Player{Id: s.id, Name: s.id, Coins: s.coins, Alive: true, MoveTime: i + 1}
		for idx, c := range s.cards {
			p.Cards = append(p.Cards, structs.Card{Name: c, Index: idx})
		}
		room.Players[s.id] = p
		room.AddConnected(s.id) // sem Redis, todo assento conta como conectado
	}
	room.GameEvent = structs.NewGameEvent(seats[0].id, structs.EventWaitingFirstAction, time.Now().Add(time.Minute), nil)
	return &table{t: t, room: room}
}

// roundTrip imita o Redis: grava a sala em JSON e lê de volta.
func (tb *table) roundTrip() {
	tb.t.Helper()
	data, err := json.Marshal(tb.room)
	if err != nil {
		tb.t.Fatalf("serializar sala: %v", err)
	}
	var fresh rooms.Room
	if err := json.Unmarshal(data, &fresh); err != nil {
		tb.t.Fatalf("ler sala: %v", err)
	}
	tb.room = &fresh
}

// play manda a mensagem crua do cliente; false quando o servidor recusa (parse ou validação).
func (tb *table) play(playerID, raw string) bool {
	tb.t.Helper()
	tb.roundTrip()
	play, err := structs.ParsePlayerPlay([]byte(raw), playerID)
	if err != nil {
		return false
	}
	return engine.ApplyPlay(context.Background(), nil, tb.room, play, registry)
}

func (tb *table) mustPlay(playerID, raw string) {
	tb.t.Helper()
	if !tb.play(playerID, raw) {
		tb.t.Fatalf("%s: jogada recusada: %s (evento %s)", playerID, raw, tb.eventType())
	}
}

func (tb *table) mustReject(playerID, raw string) {
	tb.t.Helper()
	if tb.play(playerID, raw) {
		tb.t.Fatalf("%s: jogada deveria ser recusada: %s", playerID, raw)
	}
}

// tick = a janela atual venceu (ninguém mais reagiu).
func (tb *table) tick() {
	tb.t.Helper()
	tb.roundTrip()
	engine.Tick(context.Background(), nil, registry, tb.room)
}

// tickUntil avança até abrir uma janela do tipo pedido; falha se a partida passar dela.
func (tb *table) tickUntil(want structs.GameEventType) {
	tb.t.Helper()
	for range 50 {
		if tb.eventType() == want {
			return
		}
		tb.tick()
	}
	tb.t.Fatalf("janela %s nunca abriu (última: %s)", want, tb.eventType())
}

// settle deixa tudo resolver até a próxima primeira ação (ou fim de jogo).
func (tb *table) settle() {
	tb.t.Helper()
	tb.tick()
	for range 50 {
		if tb.room.GameOver || tb.eventType() == structs.EventWaitingFirstAction {
			return
		}
		tb.tick()
	}
	tb.t.Fatalf("a sala não estabilizou (evento %s)", tb.eventType())
}

func (tb *table) eventType() structs.GameEventType {
	if tb.room.GameEvent == nil {
		return ""
	}
	return tb.room.GameEvent.Type
}

func (tb *table) p(id string) *players.Player {
	tb.t.Helper()
	p, ok := tb.room.Players[id]
	if !ok {
		tb.t.Fatalf("jogador %s não existe", id)
	}
	return p
}

func (tb *table) wantCoins(id string, want int) {
	tb.t.Helper()
	if got := tb.p(id).Coins; got != want {
		tb.t.Fatalf("%s tem %d moedas, quer %d", id, got, want)
	}
}

func (tb *table) wantDead(id string, index int, dead bool) {
	tb.t.Helper()
	if got := tb.p(id).Cards[index].Dead; got != dead {
		tb.t.Fatalf("%s carta %d morta=%v, quer %v", id, index, got, dead)
	}
}
