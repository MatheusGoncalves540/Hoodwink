// Package runner reproduz uma partida a partir de um replay .hwr (formato),
// usando o motor congelado da versão gravada no arquivo (engines). Não depende
// de Redis nem de websocket: o Sink é injetado e, por padrão, é um no-op.
//
// A sala circula como `any`: o runner não conhece os tipos congelados de cada
// versão, só a interface engines.Engine, então uma versão nova não exige mexer
// aqui (só regenerar engines/registry.go com `make engine-bump`).
package runner

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines"
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
	// Sink é o ponto de saída do motor (o Sink congelado da versão da partida).
	// Quando nil, o motor usa o no-op da própria versão.
	Sink any
	// MaxSteps é o teto de passos (padrão 5000).
	MaxSteps int
}

// Runner reproduz um replay passo a passo, mantendo a sala viva e pontos de
// retorno a cada 20 passos para a barra de tempo.
type Runner struct {
	replay   *format.Replay
	engine   engines.Engine
	sink     any
	maxSteps int

	start time.Time

	room      any
	pos       int
	elapsedMs int64

	checkpointPositions []int
	nextCheckpointIdx   int

	restorePoints []restorePoint

	closed bool
}

type restorePoint struct {
	pos       int
	room      any
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
	if err := eng.CheckMode(replay.Mode); err != nil {
		return nil, err
	}

	if len(replay.Entries) == 0 {
		return nil, errors.New("replay sem entradas")
	}
	if replay.Entries[0].Type != format.EntryStart {
		return nil, errors.New("primeira entrada não é o início")
	}

	maxSteps := opts.MaxSteps
	if maxSteps <= 0 {
		maxSteps = defaultMaxSteps
	}

	return &Runner{
		replay:              replay,
		engine:              eng,
		sink:                opts.Sink,
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
		got, err := r.checkpointOf(r.room)
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
		room = r.engine.Clone(room)
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
	var room any
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

// State devolve a sala atual (nil antes do início), no tipo congelado da versão.
func (r *Runner) State() any { return r.room }

// GameOver diz se a partida terminou.
func (r *Runner) GameOver() bool { return r.room != nil && r.engine.GameOver(r.room) }

// WinnerID devolve o id do vencedor no estado atual ("" antes do fim).
func (r *Runner) WinnerID() string {
	if r.room == nil {
		return ""
	}
	return r.engine.WinnerID(r.room)
}

// Position devolve quantas entradas já foram aplicadas.
func (r *Runner) Position() int { return r.pos }

// Total devolve o número de entradas do replay.
func (r *Runner) Total() int { return len(r.replay.Entries) }

// Hash devolve o hash canônico do estado atual (sha256 do JSON da sala), em hex.
// Vazio antes do início.
func (r *Runner) Hash() string {
	h, err := r.canonicalHash(r.room)
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
		room, err := r.engine.NewRoom(r.replay)
		if err != nil {
			return false, err
		}
		r.engine.SetNow(room, now)
		if e.Connected != nil {
			r.engine.SetConnected(room, connectedIDs(r.replay, e.Connected))
		}
		if err := r.engine.StartGame(room, r.sink); err != nil {
			r.engine.ClearNow(room)
			return false, err
		}
		r.engine.ClearNow(room)
		r.room = room
		return true, nil
	}

	if r.room == nil {
		return false, fmt.Errorf("entrada antes do início no passo %d", r.pos)
	}

	work := r.engine.Clone(r.room)
	r.engine.SetNow(work, now)

	if e.Connected != nil {
		r.engine.SetConnected(work, connectedIDs(r.replay, e.Connected))
	}

	switch e.Type {
	case format.EntryPlay:
		accepted, err := r.engine.ApplyPlay(work, r.sink, engines.PlayerID(e.Player), e.Payload)
		if err != nil {
			r.engine.ClearNow(work)
			return false, err
		}
		if accepted {
			r.room = work
		}
		r.engine.ClearNow(work)
		return accepted, nil

	case format.EntryTick:
		if err := r.engine.Tick(work, r.sink); err != nil {
			r.engine.ClearNow(work)
			return false, err
		}
		r.room = work
		r.engine.ClearNow(work)
		return true, nil

	case format.EntryConnect, format.EntryDisconnect, format.EntryJoin, format.EntryLeave:
		r.room = work
		r.engine.ClearNow(work)
		return true, nil

	case format.EntryEnd:
		if !r.engine.GameOver(work) {
			r.engine.ClearNow(work)
			return false, fmt.Errorf("fim sem GameOver no passo %d", r.pos)
		}
		wantWinner := ""
		if e.Player != format.NoPlayer {
			wantWinner = engines.PlayerID(e.Player)
		}
		if got := r.engine.WinnerID(work); got != wantWinner {
			r.engine.ClearNow(work)
			return false, fmt.Errorf("vencedor divergente no passo %d: %q != %q", r.pos, got, wantWinner)
		}
		r.room = work
		r.engine.ClearNow(work)
		return true, nil

	default:
		r.engine.ClearNow(work)
		return false, fmt.Errorf("tipo de entrada desconhecido no passo %d: %d", r.pos, e.Type)
	}
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

// checkpointOf devolve os primeiros 8 bytes do sha256 do estado canônico da sala.
func (r *Runner) checkpointOf(room any) (format.Checkpoint, error) {
	if room == nil {
		return format.Checkpoint{}, errors.New("sala não iniciada")
	}
	b, err := r.engine.CanonicalJSON(room)
	if err != nil {
		return format.Checkpoint{}, err
	}
	sum := sha256.Sum256(b)
	var c format.Checkpoint
	copy(c[:], sum[:format.CheckpointSize])
	return c, nil
}

// canonicalHash devolve o sha256 completo do estado canônico da sala, em hex.
func (r *Runner) canonicalHash(room any) (string, error) {
	if room == nil {
		return "", nil
	}
	b, err := r.engine.CanonicalJSON(room)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func connectedIDs(replay *format.Replay, bitmap []byte) []string {
	ids := make([]string, 0, len(replay.Players))
	for _, p := range replay.Players {
		if len(bitmap) > int(p.Index)/8 && bitmap[int(p.Index)/8]&(1<<(p.Index%8)) != 0 {
			ids = append(ids, engines.PlayerID(p.Index))
		}
	}
	return ids
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
