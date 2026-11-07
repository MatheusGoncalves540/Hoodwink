// Code generated pelo enginebump (make engine-bump). NÃO edite à mão.
//
// Package engines registra as cópias congeladas do motor por versão.
// O runner consulta Registry e conversa com a sala como any: cada versão
// faz o cast para os seus próprios tipos congelados.
package engines

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	enginev1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/engine"
	effectsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/engine/effects"
	rulesv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	structsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	playersv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/players"
	roomsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
	enginev2 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/engine"
	effectsv2 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/engine/effects"
	rulesv2 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/rules"
	structsv2 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs"
	playersv2 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/players"
	roomsv2 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/rooms"

	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
)

// Engine isola o runner dos tipos congelados de cada versão. A sala circula
// como any; cada versão faz o cast para os seus próprios tipos.
type Engine interface {
	Version() int
	NewRoom(*format.Replay) (any, error)
	Clone(any) any
	SetNow(any, time.Time)
	ClearNow(any)
	SetConnected(any, []string)
	StartGame(raw, sink any) error
	ApplyPlay(raw, sink any, player string, payload []byte) (bool, error)
	Tick(raw, sink any) error
	CheckMode(string) error
	GameOver(any) bool
	WinnerID(any) string
	CanonicalJSON(any) ([]byte, error)
}

// matchIDString formata os 16 bytes do MatchID como UUID canônico.
func matchIDString(b [16]byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// PlayerID mapeia o índice de jogador do replay para o id usado na sala.
func PlayerID(index uint8) string { return strconv.Itoa(int(index)) }

func truncateForHash(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.Truncate(time.Millisecond)
}

// EngineV1 liga as funções do motor v1 aos tipos congelados dessa versão.
// O runner só conversa pela interface Engine, com a sala como any.
type EngineV1 struct{}

// Version identifica a versão deste motor.
func (EngineV1) Version() int { return 1 }

var rulesV1 struct {
	once     sync.Once
	registry *rulesv1.Registry
	err      error
}

func loadRulesV1() (*rulesv1.Registry, error) {
	rulesV1.once.Do(func() {
		rulesV1.registry, rulesV1.err = rulesv1.LoadEmbeddedRegistry()
	})
	return rulesV1.registry, rulesV1.err
}

func sinkV1(raw any) roomsv1.Sink {
	if s, ok := raw.(roomsv1.Sink); ok && s != nil {
		return s
	}
	return roomsv1.NoopSink()
}

func roomV1(raw any) *roomsv1.Room { return raw.(*roomsv1.Room) }

// NewRoom monta a sala inicial a partir do cabeçalho do replay. O baralho é
// embaralhado com o RNG da sala (semeado), então o estado inicial é determinístico.
func (EngineV1) NewRoom(replay *format.Replay) (any, error) {
	registry, err := loadRulesV1()
	if err != nil {
		return nil, err
	}
	seed := roomsv1.Seed(replay.Seed)
	matchID := matchIDString(replay.MatchID)

	room := &roomsv1.Room{
		ID:          matchID,
		Rules:       structsv1.Rules(replay.Mode),
		TimeoutType: string(rulesv1.TimeoutsTypeDefault),
		Turn:        1,
		DoubledCardValues: map[structsv1.TypePlayerPlays]structsv1.DoubledCardValues{
			structsv1.PlayPoliticalCard: {},
			structsv1.PlayRebelCard:     {},
		},
		Players:                   map[string]*playersv1.Player{},
		GameEvent:                 structsv1.NewGameEvent("", structsv1.EventWaitingStart, time.Time{}, nil),
		PendingEffects:            []structsv1.EffectDTO{},
		PendingPresentationEvents: []structsv1.PresentationEventDTO{},
		Created:                   time.UnixMilli(replay.StartedAtMs).UTC(),
		MatchID:                   matchID,
		Seed:                      seed,
		Rng:                       roomsv1.NewRngState(seed),
	}

	general, err := room.GetGeneralRules(registry)
	if err != nil {
		return nil, err
	}
	if general.MaxPlayers != nil {
		room.MaxPlayers = *general.MaxPlayers
	}
	if general.InitialCoins == nil {
		return nil, errors.New("regras sem initialCoins")
	}
	if general.CopysOfEachCard == nil {
		return nil, errors.New("regras sem copysOfEachCard")
	}

	deck := make([]structsv1.CardName, 0, *general.CopysOfEachCard*len(structsv1.AllCards))
	for i := 0; i < *general.CopysOfEachCard; i++ {
		deck = append(deck, structsv1.AllCards...)
	}
	room.Rand().Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	room.Deck = deck

	for i, p := range replay.Players {
		id := PlayerID(p.Index)
		room.Players[id] = &playersv1.Player{
			Id:       id,
			Name:     p.Name,
			MoveTime: i,
			Cards: []structsv1.Card{
				{Name: structsv1.CardKamikaze, Index: 0},
				{Name: structsv1.CardKamikaze, Index: 1},
			},
			Coins: *general.InitialCoins,
			Alive: true,
		}
		room.NextMoveTime++
		room.AddConnected(id)
	}

	return room, nil
}

func (EngineV1) Clone(raw any) any { return roomV1(raw).Clone() }

func (EngineV1) SetNow(raw any, now time.Time) { roomV1(raw).SetNow(now) }

func (EngineV1) ClearNow(raw any) { roomV1(raw).ClearNow() }

func (EngineV1) SetConnected(raw any, ids []string) { roomV1(raw).SetConnected(ids) }

func (EngineV1) StartGame(raw, sink any) error {
	return effectsv1.StartGameEffect(context.Background(), sinkV1(sink), roomV1(raw))
}

func (EngineV1) ApplyPlay(raw, sink any, player string, payload []byte) (bool, error) {
	registry, err := loadRulesV1()
	if err != nil {
		return false, err
	}
	play, err := structsv1.ParsePlayerPlay(payload, player)
	if err != nil {
		return false, err
	}
	return enginev1.ApplyPlay(context.Background(), sinkV1(sink), roomV1(raw), play, registry), nil
}

func (EngineV1) Tick(raw, sink any) error {
	registry, err := loadRulesV1()
	if err != nil {
		return err
	}
	enginev1.Tick(context.Background(), sinkV1(sink), registry, roomV1(raw))
	return nil
}

func (EngineV1) CheckMode(mode string) error {
	registry, err := loadRulesV1()
	if err != nil {
		return err
	}
	if _, err := registry.Get(mode); err != nil {
		return fmt.Errorf("modo desconhecido %q: %w", mode, err)
	}
	return nil
}

func (EngineV1) GameOver(raw any) bool { return roomV1(raw).GameOver }

func (EngineV1) WinnerID(raw any) string { return roomV1(raw).WinnerID }

func (EngineV1) StateJSON(raw any) ([]byte, error) { return json.Marshal(raw) }

func (EngineV1) CanonicalJSON(raw any) ([]byte, error) { return canonicalJSONV1(roomV1(raw)) }

// normalizedV1 devolve a mesma canonicalização do gravador: campos de
// exibição zerados, IDs de jogador na ordem lexicográfica (aqui já são os
// índices), MoveTime denso, tempos truncados em ms.
func normalizedV1(room *roomsv1.Room) *roomsv1.Room {
	clone := room.Clone()
	clone.ID = clone.MatchID
	clone.Name = ""
	clone.Password = ""
	clone.MaxPlayers = 0
	clone.CustomMatch = false
	clone.Chat = nil
	clone.NonReplayable = false

	ids := make([]string, 0, len(clone.Players))
	for id := range clone.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	moveOrder := append([]string(nil), ids...)
	sort.Slice(moveOrder, func(i, j int) bool {
		a, b := clone.Players[moveOrder[i]], clone.Players[moveOrder[j]]
		if a.MoveTime != b.MoveTime {
			return a.MoveTime < b.MoveTime
		}
		return moveOrder[i] < moveOrder[j]
	})
	moveRank := make(map[string]int, len(ids))
	for i, id := range moveOrder {
		moveRank[id] = i
	}

	players := make(map[string]*playersv1.Player, len(ids))
	for id, p := range clone.Players {
		p.Name = ""
		p.Id = id
		p.MoveTime = moveRank[id]
		players[id] = p
	}
	clone.Players = players
	clone.NextMoveTime = len(ids)

	clone.StartTime = truncateForHash(clone.StartTime)
	clone.Created = clone.StartTime
	if clone.GameEvent != nil {
		clone.GameEvent.ExpiresAt = truncateForHash(clone.GameEvent.ExpiresAt)
	}

	return clone
}

// canonicalJSONV1 serializa a sala canônica passando por map[string]any, para
// os mapas saírem com chaves ordenadas exatamente como no gravador.
func canonicalJSONV1(room *roomsv1.Room) ([]byte, error) {
	data, err := json.Marshal(normalizedV1(room))
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// EngineV2 liga as funções do motor v2 aos tipos congelados dessa versão.
// O runner só conversa pela interface Engine, com a sala como any.
type EngineV2 struct{}

// Version identifica a versão deste motor.
func (EngineV2) Version() int { return 2 }

var rulesV2 struct {
	once     sync.Once
	registry *rulesv2.Registry
	err      error
}

func loadRulesV2() (*rulesv2.Registry, error) {
	rulesV2.once.Do(func() {
		rulesV2.registry, rulesV2.err = rulesv2.LoadEmbeddedRegistry()
	})
	return rulesV2.registry, rulesV2.err
}

func sinkV2(raw any) roomsv2.Sink {
	if s, ok := raw.(roomsv2.Sink); ok && s != nil {
		return s
	}
	return roomsv2.NoopSink()
}

func roomV2(raw any) *roomsv2.Room { return raw.(*roomsv2.Room) }

// NewRoom monta a sala inicial a partir do cabeçalho do replay. O baralho é
// embaralhado com o RNG da sala (semeado), então o estado inicial é determinístico.
func (EngineV2) NewRoom(replay *format.Replay) (any, error) {
	registry, err := loadRulesV2()
	if err != nil {
		return nil, err
	}
	seed := roomsv2.Seed(replay.Seed)
	matchID := matchIDString(replay.MatchID)

	room := &roomsv2.Room{
		ID:          matchID,
		Rules:       structsv2.Rules(replay.Mode),
		TimeoutType: string(rulesv2.TimeoutsTypeDefault),
		Turn:        1,
		DoubledCardValues: map[structsv2.TypePlayerPlays]structsv2.DoubledCardValues{
			structsv2.PlayPoliticalCard: {},
			structsv2.PlayRebelCard:     {},
		},
		Players:                   map[string]*playersv2.Player{},
		GameEvent:                 structsv2.NewGameEvent("", structsv2.EventWaitingStart, time.Time{}, nil),
		PendingEffects:            []structsv2.EffectDTO{},
		PendingPresentationEvents: []structsv2.PresentationEventDTO{},
		Created:                   time.UnixMilli(replay.StartedAtMs).UTC(),
		MatchID:                   matchID,
		Seed:                      seed,
		Rng:                       roomsv2.NewRngState(seed),
	}

	general, err := room.GetGeneralRules(registry)
	if err != nil {
		return nil, err
	}
	if general.MaxPlayers != nil {
		room.MaxPlayers = *general.MaxPlayers
	}
	if general.InitialCoins == nil {
		return nil, errors.New("regras sem initialCoins")
	}
	if general.CopysOfEachCard == nil {
		return nil, errors.New("regras sem copysOfEachCard")
	}

	deck := make([]structsv2.CardName, 0, *general.CopysOfEachCard*len(structsv2.AllCards))
	for i := 0; i < *general.CopysOfEachCard; i++ {
		deck = append(deck, structsv2.AllCards...)
	}
	room.Rand().Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	room.Deck = deck

	for i, p := range replay.Players {
		id := PlayerID(p.Index)
		room.Players[id] = &playersv2.Player{
			Id:       id,
			Name:     p.Name,
			MoveTime: i,
			Cards: []structsv2.Card{
				{Name: structsv2.CardKamikaze, Index: 0},
				{Name: structsv2.CardKamikaze, Index: 1},
			},
			Coins: *general.InitialCoins,
			Alive: true,
		}
		room.NextMoveTime++
		room.AddConnected(id)
	}

	return room, nil
}

func (EngineV2) Clone(raw any) any { return roomV2(raw).Clone() }

func (EngineV2) SetNow(raw any, now time.Time) { roomV2(raw).SetNow(now) }

func (EngineV2) ClearNow(raw any) { roomV2(raw).ClearNow() }

func (EngineV2) SetConnected(raw any, ids []string) { roomV2(raw).SetConnected(ids) }

func (EngineV2) StartGame(raw, sink any) error {
	return effectsv2.StartGameEffect(context.Background(), sinkV2(sink), roomV2(raw))
}

func (EngineV2) ApplyPlay(raw, sink any, player string, payload []byte) (bool, error) {
	registry, err := loadRulesV2()
	if err != nil {
		return false, err
	}
	play, err := structsv2.ParsePlayerPlay(payload, player)
	if err != nil {
		return false, err
	}
	return enginev2.ApplyPlay(context.Background(), sinkV2(sink), roomV2(raw), play, registry), nil
}

func (EngineV2) Tick(raw, sink any) error {
	registry, err := loadRulesV2()
	if err != nil {
		return err
	}
	enginev2.Tick(context.Background(), sinkV2(sink), registry, roomV2(raw))
	return nil
}

func (EngineV2) CheckMode(mode string) error {
	registry, err := loadRulesV2()
	if err != nil {
		return err
	}
	if _, err := registry.Get(mode); err != nil {
		return fmt.Errorf("modo desconhecido %q: %w", mode, err)
	}
	return nil
}

func (EngineV2) GameOver(raw any) bool { return roomV2(raw).GameOver }

func (EngineV2) WinnerID(raw any) string { return roomV2(raw).WinnerID }

func (EngineV2) StateJSON(raw any) ([]byte, error) { return json.Marshal(raw) }

func (EngineV2) CanonicalJSON(raw any) ([]byte, error) { return canonicalJSONV2(roomV2(raw)) }

// normalizedV2 devolve a mesma canonicalização do gravador: campos de
// exibição zerados, IDs de jogador na ordem lexicográfica (aqui já são os
// índices), MoveTime denso, tempos truncados em ms.
func normalizedV2(room *roomsv2.Room) *roomsv2.Room {
	clone := room.Clone()
	clone.ID = clone.MatchID
	clone.Name = ""
	clone.Password = ""
	clone.MaxPlayers = 0
	clone.CustomMatch = false
	clone.Chat = nil
	clone.NonReplayable = false

	ids := make([]string, 0, len(clone.Players))
	for id := range clone.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	moveOrder := append([]string(nil), ids...)
	sort.Slice(moveOrder, func(i, j int) bool {
		a, b := clone.Players[moveOrder[i]], clone.Players[moveOrder[j]]
		if a.MoveTime != b.MoveTime {
			return a.MoveTime < b.MoveTime
		}
		return moveOrder[i] < moveOrder[j]
	})
	moveRank := make(map[string]int, len(ids))
	for i, id := range moveOrder {
		moveRank[id] = i
	}

	players := make(map[string]*playersv2.Player, len(ids))
	for id, p := range clone.Players {
		p.Name = ""
		p.Id = id
		p.MoveTime = moveRank[id]
		players[id] = p
	}
	clone.Players = players
	clone.NextMoveTime = len(ids)

	clone.StartTime = truncateForHash(clone.StartTime)
	clone.Created = clone.StartTime
	if clone.GameEvent != nil {
		clone.GameEvent.ExpiresAt = truncateForHash(clone.GameEvent.ExpiresAt)
	}

	return clone
}

// canonicalJSONV2 serializa a sala canônica passando por map[string]any, para
// os mapas saírem com chaves ordenadas exatamente como no gravador.
func canonicalJSONV2(room *roomsv2.Room) ([]byte, error) {
	data, err := json.Marshal(normalizedV2(room))
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// Registry associa o número da versão ao motor congelado.
var Registry = map[int]Engine{
	1: EngineV1{},
	2: EngineV2{},
}
