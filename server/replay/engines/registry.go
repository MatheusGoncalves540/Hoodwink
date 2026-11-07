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
	enginev3 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/engine"
	effectsv3 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/engine/effects"
	rulesv3 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	structsv3 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	playersv3 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/players"
	roomsv3 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
	enginev4 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/engine"
	effectsv4 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/engine/effects"
	rulesv4 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	structsv4 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	playersv4 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/players"
	roomsv4 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
	enginev5 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/engine"
	effectsv5 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/engine/effects"
	rulesv5 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	structsv5 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	playersv5 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/players"
	roomsv5 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"

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

// EngineV3 liga as funções do motor v3 aos tipos congelados dessa versão.
// O runner só conversa pela interface Engine, com a sala como any.
type EngineV3 struct{}

// Version identifica a versão deste motor.
func (EngineV3) Version() int { return 3 }

var rulesV3 struct {
	once     sync.Once
	registry *rulesv3.Registry
	err      error
}

func loadRulesV3() (*rulesv3.Registry, error) {
	rulesV3.once.Do(func() {
		rulesV3.registry, rulesV3.err = rulesv3.LoadEmbeddedRegistry()
	})
	return rulesV3.registry, rulesV3.err
}

func sinkV3(raw any) roomsv3.Sink {
	if s, ok := raw.(roomsv3.Sink); ok && s != nil {
		return s
	}
	return roomsv3.NoopSink()
}

func roomV3(raw any) *roomsv3.Room { return raw.(*roomsv3.Room) }

// NewRoom monta a sala inicial a partir do cabeçalho do replay. O baralho é
// embaralhado com o RNG da sala (semeado), então o estado inicial é determinístico.
func (EngineV3) NewRoom(replay *format.Replay) (any, error) {
	registry, err := loadRulesV3()
	if err != nil {
		return nil, err
	}
	seed := roomsv3.Seed(replay.Seed)
	matchID := matchIDString(replay.MatchID)

	room := &roomsv3.Room{
		ID:          matchID,
		Rules:       structsv3.Rules(replay.Mode),
		TimeoutType: string(rulesv3.TimeoutsTypeDefault),
		Turn:        1,
		DoubledCardValues: map[structsv3.TypePlayerPlays]structsv3.DoubledCardValues{
			structsv3.PlayPoliticalCard: {},
			structsv3.PlayRebelCard:     {},
		},
		Players:                   map[string]*playersv3.Player{},
		GameEvent:                 structsv3.NewGameEvent("", structsv3.EventWaitingStart, time.Time{}, nil),
		PendingEffects:            []structsv3.EffectDTO{},
		PendingPresentationEvents: []structsv3.PresentationEventDTO{},
		Created:                   time.UnixMilli(replay.StartedAtMs).UTC(),
		MatchID:                   matchID,
		Seed:                      seed,
		Rng:                       roomsv3.NewRngState(seed),
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

	deck := make([]structsv3.CardName, 0, *general.CopysOfEachCard*len(structsv3.AllCards))
	for i := 0; i < *general.CopysOfEachCard; i++ {
		deck = append(deck, structsv3.AllCards...)
	}
	room.Rand().Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	room.Deck = deck

	for i, p := range replay.Players {
		id := PlayerID(p.Index)
		room.Players[id] = &playersv3.Player{
			Id:       id,
			Name:     p.Name,
			MoveTime: i,
			Cards: []structsv3.Card{
				{Name: structsv3.CardKamikaze, Index: 0},
				{Name: structsv3.CardKamikaze, Index: 1},
			},
			Coins: *general.InitialCoins,
			Alive: true,
		}
		room.NextMoveTime++
		room.AddConnected(id)
	}

	return room, nil
}

func (EngineV3) Clone(raw any) any { return roomV3(raw).Clone() }

func (EngineV3) SetNow(raw any, now time.Time) { roomV3(raw).SetNow(now) }

func (EngineV3) ClearNow(raw any) { roomV3(raw).ClearNow() }

func (EngineV3) SetConnected(raw any, ids []string) { roomV3(raw).SetConnected(ids) }

func (EngineV3) StartGame(raw, sink any) error {
	return effectsv3.StartGameEffect(context.Background(), sinkV3(sink), roomV3(raw))
}

func (EngineV3) ApplyPlay(raw, sink any, player string, payload []byte) (bool, error) {
	registry, err := loadRulesV3()
	if err != nil {
		return false, err
	}
	play, err := structsv3.ParsePlayerPlay(payload, player)
	if err != nil {
		return false, err
	}
	return enginev3.ApplyPlay(context.Background(), sinkV3(sink), roomV3(raw), play, registry), nil
}

func (EngineV3) Tick(raw, sink any) error {
	registry, err := loadRulesV3()
	if err != nil {
		return err
	}
	enginev3.Tick(context.Background(), sinkV3(sink), registry, roomV3(raw))
	return nil
}

func (EngineV3) CheckMode(mode string) error {
	registry, err := loadRulesV3()
	if err != nil {
		return err
	}
	if _, err := registry.Get(mode); err != nil {
		return fmt.Errorf("modo desconhecido %q: %w", mode, err)
	}
	return nil
}

func (EngineV3) GameOver(raw any) bool { return roomV3(raw).GameOver }

func (EngineV3) WinnerID(raw any) string { return roomV3(raw).WinnerID }

func (EngineV3) StateJSON(raw any) ([]byte, error) { return json.Marshal(raw) }

func (EngineV3) CanonicalJSON(raw any) ([]byte, error) { return canonicalJSONV3(roomV3(raw)) }

// normalizedV3 devolve a mesma canonicalização do gravador: campos de
// exibição zerados, IDs de jogador na ordem lexicográfica (aqui já são os
// índices), MoveTime denso, tempos truncados em ms.
func normalizedV3(room *roomsv3.Room) *roomsv3.Room {
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

	players := make(map[string]*playersv3.Player, len(ids))
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

// canonicalJSONV3 serializa a sala canônica passando por map[string]any, para
// os mapas saírem com chaves ordenadas exatamente como no gravador.
func canonicalJSONV3(room *roomsv3.Room) ([]byte, error) {
	data, err := json.Marshal(normalizedV3(room))
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// EngineV4 liga as funções do motor v4 aos tipos congelados dessa versão.
// O runner só conversa pela interface Engine, com a sala como any.
type EngineV4 struct{}

// Version identifica a versão deste motor.
func (EngineV4) Version() int { return 4 }

var rulesV4 struct {
	once     sync.Once
	registry *rulesv4.Registry
	err      error
}

func loadRulesV4() (*rulesv4.Registry, error) {
	rulesV4.once.Do(func() {
		rulesV4.registry, rulesV4.err = rulesv4.LoadEmbeddedRegistry()
	})
	return rulesV4.registry, rulesV4.err
}

func sinkV4(raw any) roomsv4.Sink {
	if s, ok := raw.(roomsv4.Sink); ok && s != nil {
		return s
	}
	return roomsv4.NoopSink()
}

func roomV4(raw any) *roomsv4.Room { return raw.(*roomsv4.Room) }

// NewRoom monta a sala inicial a partir do cabeçalho do replay. O baralho é
// embaralhado com o RNG da sala (semeado), então o estado inicial é determinístico.
func (EngineV4) NewRoom(replay *format.Replay) (any, error) {
	registry, err := loadRulesV4()
	if err != nil {
		return nil, err
	}
	seed := roomsv4.Seed(replay.Seed)
	matchID := matchIDString(replay.MatchID)

	room := &roomsv4.Room{
		ID:          matchID,
		Rules:       structsv4.Rules(replay.Mode),
		TimeoutType: string(rulesv4.TimeoutsTypeDefault),
		Turn:        1,
		DoubledCardValues: map[structsv4.TypePlayerPlays]structsv4.DoubledCardValues{
			structsv4.PlayPoliticalCard: {},
			structsv4.PlayRebelCard:     {},
		},
		Players:                   map[string]*playersv4.Player{},
		GameEvent:                 structsv4.NewGameEvent("", structsv4.EventWaitingStart, time.Time{}, nil),
		PendingEffects:            []structsv4.EffectDTO{},
		PendingPresentationEvents: []structsv4.PresentationEventDTO{},
		Created:                   time.UnixMilli(replay.StartedAtMs).UTC(),
		MatchID:                   matchID,
		Seed:                      seed,
		Rng:                       roomsv4.NewRngState(seed),
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

	deck := make([]structsv4.CardName, 0, *general.CopysOfEachCard*len(structsv4.AllCards))
	for i := 0; i < *general.CopysOfEachCard; i++ {
		deck = append(deck, structsv4.AllCards...)
	}
	room.Rand().Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	room.Deck = deck

	for i, p := range replay.Players {
		id := PlayerID(p.Index)
		room.Players[id] = &playersv4.Player{
			Id:       id,
			Name:     p.Name,
			MoveTime: i,
			Cards: []structsv4.Card{
				{Name: structsv4.CardKamikaze, Index: 0},
				{Name: structsv4.CardKamikaze, Index: 1},
			},
			Coins: *general.InitialCoins,
			Alive: true,
		}
		room.NextMoveTime++
		room.AddConnected(id)
	}

	return room, nil
}

func (EngineV4) Clone(raw any) any { return roomV4(raw).Clone() }

func (EngineV4) SetNow(raw any, now time.Time) { roomV4(raw).SetNow(now) }

func (EngineV4) ClearNow(raw any) { roomV4(raw).ClearNow() }

func (EngineV4) SetConnected(raw any, ids []string) { roomV4(raw).SetConnected(ids) }

func (EngineV4) StartGame(raw, sink any) error {
	return effectsv4.StartGameEffect(context.Background(), sinkV4(sink), roomV4(raw))
}

func (EngineV4) ApplyPlay(raw, sink any, player string, payload []byte) (bool, error) {
	registry, err := loadRulesV4()
	if err != nil {
		return false, err
	}
	play, err := structsv4.ParsePlayerPlay(payload, player)
	if err != nil {
		return false, err
	}
	return enginev4.ApplyPlay(context.Background(), sinkV4(sink), roomV4(raw), play, registry), nil
}

func (EngineV4) Tick(raw, sink any) error {
	registry, err := loadRulesV4()
	if err != nil {
		return err
	}
	enginev4.Tick(context.Background(), sinkV4(sink), registry, roomV4(raw))
	return nil
}

func (EngineV4) CheckMode(mode string) error {
	registry, err := loadRulesV4()
	if err != nil {
		return err
	}
	if _, err := registry.Get(mode); err != nil {
		return fmt.Errorf("modo desconhecido %q: %w", mode, err)
	}
	return nil
}

func (EngineV4) GameOver(raw any) bool { return roomV4(raw).GameOver }

func (EngineV4) WinnerID(raw any) string { return roomV4(raw).WinnerID }

func (EngineV4) StateJSON(raw any) ([]byte, error) { return json.Marshal(raw) }

func (EngineV4) CanonicalJSON(raw any) ([]byte, error) { return canonicalJSONV4(roomV4(raw)) }

// normalizedV4 devolve a mesma canonicalização do gravador: campos de
// exibição zerados, IDs de jogador na ordem lexicográfica (aqui já são os
// índices), MoveTime denso, tempos truncados em ms.
func normalizedV4(room *roomsv4.Room) *roomsv4.Room {
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

	players := make(map[string]*playersv4.Player, len(ids))
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

// canonicalJSONV4 serializa a sala canônica passando por map[string]any, para
// os mapas saírem com chaves ordenadas exatamente como no gravador.
func canonicalJSONV4(room *roomsv4.Room) ([]byte, error) {
	data, err := json.Marshal(normalizedV4(room))
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

// EngineV5 liga as funções do motor v5 aos tipos congelados dessa versão.
// O runner só conversa pela interface Engine, com a sala como any.
type EngineV5 struct{}

// Version identifica a versão deste motor.
func (EngineV5) Version() int { return 5 }

var rulesV5 struct {
	once     sync.Once
	registry *rulesv5.Registry
	err      error
}

func loadRulesV5() (*rulesv5.Registry, error) {
	rulesV5.once.Do(func() {
		rulesV5.registry, rulesV5.err = rulesv5.LoadEmbeddedRegistry()
	})
	return rulesV5.registry, rulesV5.err
}

func sinkV5(raw any) roomsv5.Sink {
	if s, ok := raw.(roomsv5.Sink); ok && s != nil {
		return s
	}
	return roomsv5.NoopSink()
}

func roomV5(raw any) *roomsv5.Room { return raw.(*roomsv5.Room) }

// NewRoom monta a sala inicial a partir do cabeçalho do replay. O baralho é
// embaralhado com o RNG da sala (semeado), então o estado inicial é determinístico.
func (EngineV5) NewRoom(replay *format.Replay) (any, error) {
	registry, err := loadRulesV5()
	if err != nil {
		return nil, err
	}
	seed := roomsv5.Seed(replay.Seed)
	matchID := matchIDString(replay.MatchID)

	room := &roomsv5.Room{
		ID:          matchID,
		Rules:       structsv5.Rules(replay.Mode),
		TimeoutType: string(rulesv5.TimeoutsTypeDefault),
		Turn:        1,
		DoubledCardValues: map[structsv5.TypePlayerPlays]structsv5.DoubledCardValues{
			structsv5.PlayPoliticalCard: {},
			structsv5.PlayRebelCard:     {},
		},
		Players:                   map[string]*playersv5.Player{},
		GameEvent:                 structsv5.NewGameEvent("", structsv5.EventWaitingStart, time.Time{}, nil),
		PendingEffects:            []structsv5.EffectDTO{},
		PendingPresentationEvents: []structsv5.PresentationEventDTO{},
		Created:                   time.UnixMilli(replay.StartedAtMs).UTC(),
		MatchID:                   matchID,
		Seed:                      seed,
		Rng:                       roomsv5.NewRngState(seed),
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

	deck := make([]structsv5.CardName, 0, *general.CopysOfEachCard*len(structsv5.AllCards))
	for i := 0; i < *general.CopysOfEachCard; i++ {
		deck = append(deck, structsv5.AllCards...)
	}
	room.Rand().Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	room.Deck = deck

	for i, p := range replay.Players {
		id := PlayerID(p.Index)
		room.Players[id] = &playersv5.Player{
			Id:       id,
			Name:     p.Name,
			MoveTime: i,
			Cards: []structsv5.Card{
				{Name: structsv5.CardKamikaze, Index: 0},
				{Name: structsv5.CardKamikaze, Index: 1},
			},
			Coins: *general.InitialCoins,
			Alive: true,
		}
		room.NextMoveTime++
		room.AddConnected(id)
	}

	return room, nil
}

func (EngineV5) Clone(raw any) any { return roomV5(raw).Clone() }

func (EngineV5) SetNow(raw any, now time.Time) { roomV5(raw).SetNow(now) }

func (EngineV5) ClearNow(raw any) { roomV5(raw).ClearNow() }

func (EngineV5) SetConnected(raw any, ids []string) { roomV5(raw).SetConnected(ids) }

func (EngineV5) StartGame(raw, sink any) error {
	return effectsv5.StartGameEffect(context.Background(), sinkV5(sink), roomV5(raw))
}

func (EngineV5) ApplyPlay(raw, sink any, player string, payload []byte) (bool, error) {
	registry, err := loadRulesV5()
	if err != nil {
		return false, err
	}
	play, err := structsv5.ParsePlayerPlay(payload, player)
	if err != nil {
		return false, err
	}
	return enginev5.ApplyPlay(context.Background(), sinkV5(sink), roomV5(raw), play, registry), nil
}

func (EngineV5) Tick(raw, sink any) error {
	registry, err := loadRulesV5()
	if err != nil {
		return err
	}
	enginev5.Tick(context.Background(), sinkV5(sink), registry, roomV5(raw))
	return nil
}

func (EngineV5) CheckMode(mode string) error {
	registry, err := loadRulesV5()
	if err != nil {
		return err
	}
	if _, err := registry.Get(mode); err != nil {
		return fmt.Errorf("modo desconhecido %q: %w", mode, err)
	}
	return nil
}

func (EngineV5) GameOver(raw any) bool { return roomV5(raw).GameOver }

func (EngineV5) WinnerID(raw any) string { return roomV5(raw).WinnerID }

func (EngineV5) StateJSON(raw any) ([]byte, error) { return json.Marshal(raw) }

func (EngineV5) CanonicalJSON(raw any) ([]byte, error) { return canonicalJSONV5(roomV5(raw)) }

// normalizedV5 devolve a mesma canonicalização do gravador: campos de
// exibição zerados, IDs de jogador na ordem lexicográfica (aqui já são os
// índices), MoveTime denso, tempos truncados em ms.
func normalizedV5(room *roomsv5.Room) *roomsv5.Room {
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

	players := make(map[string]*playersv5.Player, len(ids))
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

// canonicalJSONV5 serializa a sala canônica passando por map[string]any, para
// os mapas saírem com chaves ordenadas exatamente como no gravador.
func canonicalJSONV5(room *roomsv5.Room) ([]byte, error) {
	data, err := json.Marshal(normalizedV5(room))
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
	3: EngineV3{},
	4: EngineV4{},
	5: EngineV5{},
}
