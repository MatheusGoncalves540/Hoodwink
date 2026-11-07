// Package runner reproduz uma partida a partir de um replay .hwr (formato),
// usando o motor congelado da versão gravada no arquivo (engines). Não depende
// de Redis nem de websocket: o Sink é injetado e, por padrão, é um no-op.
package runner

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines"
	effectsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/engine/effects"
	rulesv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	structsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	playersv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/players"
	roomsv1 "github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
)

const (
	defaultMaxSteps = 5000
	restoreInterval = 20
)

var (
	// ErrClosed indica uso do Runner depois de Close.
	ErrClosed = errors.New("runner já foi fechado")
	// ErrDone indica que não há mais entradas para aplicar.
	ErrDone = errors.New("replay já chegou ao fim")
	// ErrTooManySteps indica que o teto de passos foi excedido.
	ErrTooManySteps = errors.New("teto de passos excedido")
)

// Options controla a execução do Runner.
type Options struct {
	// Sink é o ponto de saída do motor. Quando nil, usa um no-op (sem Redis/websocket).
	Sink roomsv1.Sink
	// Registry de regras. Quando nil, carrega o registry embutido (modos YAML).
	Registry *rulesv1.Registry
	// MaxSteps é o teto de passos (padrão 5000).
	MaxSteps int
}

// Runner reproduz um replay passo a passo, mantendo a sala viva e pontos de
// retorno a cada 20 passos para a barra de tempo.
type Runner struct {
	replay   *format.Replay
	engine   engines.EngineV1
	registry *rulesv1.Registry
	sink     roomsv1.Sink
	maxSteps int

	start time.Time

	room      *roomsv1.Room
	pos       int
	elapsedMs int64

	checkpointPositions []int
	nextCheckpointIdx   int

	restorePoints []restorePoint

	closed bool
}

type restorePoint struct {
	pos       int
	room      *roomsv1.Room
	elapsedMs int64
}

// New prepara o Runner, escolhendo a versão no registry e validando o modo.
func New(replay *format.Replay, opts Options) (*Runner, error) {
	if replay == nil {
		return nil, errors.New("replay nil")
	}

	eng, ok := engines.Registry[replay.EngineVersion]
	if !ok {
		return nil, fmt.Errorf("versão do motor desconhecida: %d", replay.EngineVersion)
	}
	v1, ok := eng.(engines.EngineV1)
	if !ok {
		return nil, fmt.Errorf("versão %d não suportada pelo runner", replay.EngineVersion)
	}

	registry := opts.Registry
	if registry == nil {
		reg, err := rulesv1.LoadEmbeddedRegistry()
		if err != nil {
			return nil, fmt.Errorf("carregar regras embutidas: %w", err)
		}
		registry = reg
	}
	if _, err := registry.Get(replay.Mode); err != nil {
		return nil, fmt.Errorf("modo desconhecido %q: %w", replay.Mode, err)
	}

	if len(replay.Entries) == 0 {
		return nil, errors.New("replay sem entradas")
	}
	if replay.Entries[0].Type != format.EntryStart {
		return nil, errors.New("primeira entrada não é o início")
	}

	sink := opts.Sink
	if sink == nil {
		sink = roomsv1.NoopSink()
	}
	maxSteps := opts.MaxSteps
	if maxSteps <= 0 {
		maxSteps = defaultMaxSteps
	}

	return &Runner{
		replay:              replay,
		engine:              v1,
		registry:            registry,
		sink:                sink,
		maxSteps:            maxSteps,
		start:               time.UnixMilli(replay.StartedAtMs).UTC(),
		checkpointPositions: checkpointPositions(len(replay.Entries)),
	}, nil
}

// Step aplica a próxima entrada e confere o checkpoint quando for o caso.
func (r *Runner) Step() error {
	if r.closed {
		return ErrClosed
	}
	if r.pos >= len(r.replay.Entries) {
		return ErrDone
	}
	if r.pos >= r.maxSteps {
		return ErrTooManySteps
	}

	if _, err := r.applyEntry(r.replay.Entries[r.pos]); err != nil {
		return err
	}
	r.pos++

	// Confere o checkpoint (a cada 20 entradas e no fim).
	if r.nextCheckpointIdx < len(r.checkpointPositions) &&
		r.pos == r.checkpointPositions[r.nextCheckpointIdx] {
		want := r.replay.Checkpoints[r.nextCheckpointIdx]
		got, err := checkpointOf(r.room)
		if err != nil {
			return err
		}
		if got != want {
			return &Divergence{Step: r.pos, Want: want, Got: got}
		}
		r.nextCheckpointIdx++
	}

	// Guarda um ponto de retorno a cada 20 passos.
	if r.pos%restoreInterval == 0 {
		r.storeRestore()
	}
	return nil
}

func (r *Runner) storeRestore() {
	room := r.room
	if room != nil {
		room = room.Clone()
	}
	r.restorePoints = append(r.restorePoints, restorePoint{pos: r.pos, room: room, elapsedMs: r.elapsedMs})
}

// Seek posiciona o Runner no passo n (0..len(entries)) reaproveitando o ponto de
// retorno mais próximo. Não confere checkpoints (é operação de navegação).
func (r *Runner) Seek(n int) error {
	if r.closed {
		return ErrClosed
	}
	if n < 0 || n > len(r.replay.Entries) {
		return fmt.Errorf("passo %d fora do intervalo 0..%d", n, len(r.replay.Entries))
	}
	if n == r.pos {
		return nil
	}

	from := 0
	var room *roomsv1.Room
	var elapsed int64
	for _, rp := range r.restorePoints {
		if rp.pos <= n {
			from = rp.pos
			room = rp.room
			elapsed = rp.elapsedMs
		}
	}

	r.pos = from
	r.room = room
	r.elapsedMs = elapsed

	for r.pos < n {
		if _, err := r.applyEntry(r.replay.Entries[r.pos]); err != nil {
			return err
		}
		r.pos++
	}

	r.nextCheckpointIdx = 0
	for r.nextCheckpointIdx < len(r.checkpointPositions) &&
		r.checkpointPositions[r.nextCheckpointIdx] <= n {
		r.nextCheckpointIdx++
	}
	return nil
}

// State devolve a sala atual (nil antes do início).
func (r *Runner) State() *roomsv1.Room { return r.room }

// Position devolve quantas entradas já foram aplicadas.
func (r *Runner) Position() int { return r.pos }

// Total devolve o número de entradas do replay.
func (r *Runner) Total() int { return len(r.replay.Entries) }

// Hash devolve o hash canônico do estado atual (sha256 do JSON da sala), em hex.
// Vazio antes do início.
func (r *Runner) Hash() string {
	h, err := canonicalHash(r.room)
	if err != nil {
		return ""
	}
	return h
}

// Close libera o Runner.
func (r *Runner) Close() error {
	if r.closed {
		return ErrClosed
	}
	r.closed = true
	r.room = nil
	r.restorePoints = nil
	return nil
}

// applyEntry aplica uma entrada à sala, com clone entre passos (ida e volta JSON)
// e relógio fixo por entrada. Recupera panic e devolve erro. start inicializa a sala.
func (r *Runner) applyEntry(e format.Entry) (applied bool, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = &StepError{Step: r.pos, Err: fmt.Errorf("panic: %v", rec)}
			applied = false
		}
	}()

	now := r.start.Add(time.Duration(r.elapsedMs+e.DeltaMs) * time.Millisecond)
	r.elapsedMs += e.DeltaMs

	if e.Type == format.EntryStart {
		room, err := newRoom(r.replay, r.registry)
		if err != nil {
			return false, err
		}
		room.SetNow(now)
		if e.Connected != nil {
			room.SetConnected(connectedIDs(r.replay, e.Connected))
		}
		if err := effectsv1.StartGameEffect(context.Background(), r.sink, room); err != nil {
			room.ClearNow()
			return false, err
		}
		room.ClearNow()
		r.room = room
		return true, nil
	}

	if r.room == nil {
		return false, fmt.Errorf("entrada antes do início no passo %d", r.pos)
	}

	work := r.room.Clone()
	work.SetNow(now)

	if e.Connected != nil {
		work.SetConnected(connectedIDs(r.replay, e.Connected))
	}

	switch e.Type {
	case format.EntryPlay:
		play, err := structsv1.ParsePlayerPlay(e.Payload, playerID(e.Player))
		if err != nil {
			work.ClearNow()
			return false, err
		}
		accepted := r.engine.ApplyPlay(context.Background(), r.sink, work, play, r.registry)
		if accepted {
			r.room = work
		}
		work.ClearNow()
		return accepted, nil

	case format.EntryTick:
		r.engine.Tick(context.Background(), r.sink, r.registry, work)
		r.room = work
		work.ClearNow()
		return true, nil

	case format.EntryConnect, format.EntryDisconnect, format.EntryJoin, format.EntryLeave:
		r.room = work
		work.ClearNow()
		return true, nil

	case format.EntryEnd:
		if !work.GameOver {
			work.ClearNow()
			return false, fmt.Errorf("fim sem GameOver no passo %d", r.pos)
		}
		wantWinner := ""
		if e.Player != format.NoPlayer {
			wantWinner = playerID(e.Player)
		}
		if work.WinnerID != wantWinner {
			work.ClearNow()
			return false, fmt.Errorf("vencedor divergente no passo %d: %q != %q", r.pos, work.WinnerID, wantWinner)
		}
		r.room = work
		work.ClearNow()
		return true, nil

	default:
		work.ClearNow()
		return false, fmt.Errorf("tipo de entrada desconhecido no passo %d: %d", r.pos, e.Type)
	}
}

// newRoom monta a sala inicial a partir do cabeçalho do replay. O baralho é
// embaralhado com o RNG da sala (semeado), então o estado inicial é determinístico.
func newRoom(replay *format.Replay, registry *rulesv1.Registry) (*roomsv1.Room, error) {
	seed := roomsv1.Seed(replay.Seed)
	matchID := matchIDString(replay.MatchID)

	room := &roomsv1.Room{
		ID:          matchID,
		Rules:       structsv1.Rules(replay.Mode),
		TimeoutType: string(rulesv1.TimeoutsTypeDefault),
		Name:        "",
		Password:    "",
		MaxPlayers:  0,
		CustomMatch: false,
		Turn:        1,
		Tax:         0,
		DoubledCardValues: map[structsv1.TypePlayerPlays]structsv1.DoubledCardValues{
			structsv1.PlayPoliticalCard: {},
			structsv1.PlayRebelCard:     {},
		},
		Players:                   map[string]*playersv1.Player{},
		Deck:                      nil,
		CurrentPlayer:             "",
		GameEvent:                 structsv1.NewGameEvent("", structsv1.EventWaitingStart, time.Time{}, nil),
		PendingEffects:            []structsv1.EffectDTO{},
		PendingPresentationEvents: []structsv1.PresentationEventDTO{},
		GameOver:                  false,
		StartTime:                 time.Time{},
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
		id := playerID(p.Index)
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

// checkpointPositions devolve as posições (1-based) em que um checkpoint é
// verificado: a cada 20 entradas e no fim.
func checkpointPositions(n int) []int {
	pos := make([]int, 0, (n+format.CheckpointInterval-1)/format.CheckpointInterval)
	for p := format.CheckpointInterval; p < n; p += format.CheckpointInterval {
		pos = append(pos, p)
	}
	pos = append(pos, n)
	return pos
}

// normalizedForHash devolve a mesma canonicalização do gravador: campos de
// exibição zerados, IDs de jogador na ordem lexicográfica (aqui já são os
// índices), MoveTime denso, tempos truncados em ms.
func normalizedForHash(room *roomsv1.Room) *roomsv1.Room {
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

func truncateForHash(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return t.Truncate(time.Millisecond)
}

// checkpointOf devolve os primeiros 8 bytes do sha256 do estado canônico da sala.
func checkpointOf(room *roomsv1.Room) (format.Checkpoint, error) {
	if room == nil {
		return format.Checkpoint{}, errors.New("sala não iniciada")
	}
	b, err := canonicalJSON(room)
	if err != nil {
		return format.Checkpoint{}, err
	}
	sum := sha256.Sum256(b)
	var c format.Checkpoint
	copy(c[:], sum[:format.CheckpointSize])
	return c, nil
}

// canonicalHash devolve o sha256 completo do estado canônico da sala, em hex.
func canonicalHash(room *roomsv1.Room) (string, error) {
	if room == nil {
		return "", nil
	}
	b, err := canonicalJSON(room)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

// canonicalJSON serializa a sala canônica passando por map[string]any, para os
// mapas saírem com chaves ordenadas exatamente como no gravador (ReplayCheckpoint).
func canonicalJSON(room *roomsv1.Room) ([]byte, error) {
	data, err := json.Marshal(normalizedForHash(room))
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func connectedIDs(replay *format.Replay, bitmap []byte) []string {
	ids := make([]string, 0, len(replay.Players))
	for _, p := range replay.Players {
		if len(bitmap) > int(p.Index)/8 && bitmap[int(p.Index)/8]&(1<<(p.Index%8)) != 0 {
			ids = append(ids, playerID(p.Index))
		}
	}
	return ids
}

// playerID mapeia o índice de jogador do replay para o id usado na sala.
func playerID(index uint8) string { return strconv.Itoa(int(index)) }

// matchIDString formata os 16 bytes do MatchID como UUID canônico.
func matchIDString(b [16]byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Divergence indica que o estado calculado divergiu do checkpoint do arquivo.
type Divergence struct {
	Step int
	Want format.Checkpoint
	Got  format.Checkpoint
}

func (d *Divergence) Error() string {
	return fmt.Sprintf("estado divergiu no passo %d: checkpoint %x != %x", d.Step, d.Want[:], d.Got[:])
}

// StepError envolve um erro ocorrido ao aplicar uma entrada, guardando o passo.
type StepError struct {
	Step int
	Err  error
}

func (e *StepError) Error() string { return fmt.Sprintf("passo %d: %v", e.Step, e.Err) }
func (e *StepError) Unwrap() error { return e.Err }
