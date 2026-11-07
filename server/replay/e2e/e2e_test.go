// Package e2e joga partidas pelo motor real do gameServer, grava as entradas no
// Redis pelo WithRoom, monta o .hwr com o recorder e reproduz com o runner
// congelado da v1. Cada passo do runner é comparado com o checkpoint gravado
// pelo servidor, então uma divergência aponta a entrada exata.
package e2e

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/connection"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
	"github.com/MatheusGoncalves540/Hoodwink-replay/runner"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

const (
	apiKey         = "e2e-key"
	goldenMatchID  = "0192f3a0-0000-7000-8000-000000000001"
	goldenFileName = "golden_1.hwr"
)

type uploadEnvelope struct {
	Data      string   `json:"data"`
	RoomID    string   `json:"room_id"`
	PlayerIDs []string `json:"player_ids"`
}

type env struct {
	t        *testing.T
	rdb      *redis.Client
	registry *rules.Registry
	backend  *services.BackendService
	srv      *httptest.Server

	mu       sync.Mutex
	envelope *uploadEnvelope
}

var silenceLogsOnce sync.Once

// silenceEngineLogs descarta o log padrão do motor (jogadas recusadas) para não
// afogar a saída dos testes que injetam centenas delas.
func silenceEngineLogs() {
	silenceLogsOnce.Do(func() { log.SetOutput(io.Discard) })
}

func newEnv(t *testing.T) *env {
	t.Helper()

	t.Setenv("REPLAY_ENABLED", "true")
	t.Setenv("ROOM_LOCK_TTL_MS", "3000")
	config.InstanceID = "e2e"

	// O motor loga cada jogada recusada; nas centenas de jogadas inválidas do
	// teste longo isso afogaria a saída. Os erros do teste continuam visíveis.
	silenceEngineLogs()

	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	// O heartbeat do gameServer marca a instância como viva; sem essa chave o
	// fillConnected considera todo jogador desconectado.
	if err := rdb.Set(context.Background(), "instance:"+config.InstanceID+":alive", "1", 0).Err(); err != nil {
		t.Fatalf("marcar instância viva: %v", err)
	}

	registry, err := rules.LoadEmbeddedRegistry()
	if err != nil {
		t.Fatalf("carregar regras embutidas: %v", err)
	}

	e := &env{t: t, rdb: rdb, registry: registry}
	e.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != apiKey {
			http.Error(w, "api key inválida", http.StatusUnauthorized)
			return
		}
		var body uploadEnvelope
		if err := json.NewDecoder(io.LimitReader(r.Body, 512*1024)).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		e.mu.Lock()
		e.envelope = &body
		e.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(e.srv.Close)

	t.Setenv("BACKEND_URL", e.srv.URL)
	t.Setenv("BACKEND_API_KEY", apiKey)
	e.backend = services.NewBackendService()

	return e
}

func (e *env) takeEnvelope() *uploadEnvelope {
	e.mu.Lock()
	defer e.mu.Unlock()
	env := e.envelope
	e.envelope = nil
	return env
}

type matchOptions struct {
	players        int
	label          string
	names          []string
	seed           *rooms.Seed
	matchID        string
	gapBeforeStart bool
	// invalid é quantas jogadas fora do turno injetar no total.
	invalid int
	// disconnect liga uma desconexão/reconexão no meio da partida.
	disconnect bool
	// padInvalid adiciona jogadas recusadas extras (partida longa).
	padInvalid int
	maxSteps   int
}

type matchResult struct {
	hwr       []byte
	finalHash string
	recorded  []redisFuncs.ReplayEntry
	finalRoom *rooms.Room
	envelope  uploadEnvelope
}

func fixedGoldenSeed() rooms.Seed {
	var seed rooms.Seed
	for i := range seed {
		seed[i] = byte(i*7 + 1)
	}
	return seed
}

// newUUID gera um UUID v4 sem puxar dependência nova para o módulo do replay.
func newUUID() string {
	var b [16]byte
	if _, err := cryptorand.Read(b[:]); err != nil {
		panic(fmt.Sprintf("gerar uuid: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func playerName(opts matchOptions, i int) string {
	if i < len(opts.names) && opts.names[i] != "" {
		return opts.names[i]
	}
	return fmt.Sprintf("jogador-%d", i)
}

// playMatch cria a sala, joga até o fim pelo motor real, finaliza o replay no
// backend httptest e confere o runner contra cada checkpoint gravado.
func playMatch(t *testing.T, e *env, opts matchOptions) *matchResult {
	t.Helper()
	ctx := context.Background()

	if opts.players == 0 {
		opts.players = 2
	}
	if opts.maxSteps == 0 {
		opts.maxSteps = 4000
	}
	if len(opts.names) < opts.players {
		names := make([]string, opts.players)
		copy(names, opts.names)
		opts.names = names
	}
	visual := newVisualReporter(opts.label, opts.players)

	roomSvc := services.NewRoomService(e.rdb)
	req := httptest.NewRequest(http.MethodPost, "/rooms", nil)
	room, err := roomSvc.CreateNewRoom(req, e.registry, endpointStructures.CreateRoomRequest{
		RoomName:   "Sala e2e",
		MaxPlayers: 10,
	}, false)
	if err != nil {
		t.Fatalf("criar sala: %v", err)
	}

	if opts.seed != nil || opts.matchID != "" {
		if err := overrideRoom(ctx, e, room, opts); err != nil {
			t.Fatalf("preparar sala determinística: %v", err)
		}
	}

	ids := make([]string, 0, opts.players)
	for i := 0; i < opts.players; i++ {
		id := newUUID()
		if err := addPlayer(ctx, e, room, id, playerName(opts, i)); err != nil {
			t.Fatalf("adicionar jogador %d: %v", i, err)
		}
		ids = append(ids, id)
	}

	if opts.gapBeforeStart {
		// Jogador que entra e sai antes do início: deixa buraco em MoveTime e em
		// NextMoveTime. O replay tem que densificar os dois lados.
		ghost := newUUID()
		if err := addPlayer(ctx, e, room, ghost, "fantasma"); err != nil {
			t.Fatalf("adicionar fantasma: %v", err)
		}
		if err := playerRedis.UnregisterPlayerFromRoom(ctx, e.rdb, ghost); err != nil {
			t.Fatalf("desregistrar fantasma: %v", err)
		}
		var removeErr error
		if err := redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CauseLeave, func(r *rooms.Room) bool {
			removed, err := redisFuncs.RemovePlayerFromRoom(r, ghost)
			if err != nil {
				removeErr = err
				return false
			}
			return removed
		}); err != nil {
			t.Fatalf("sair com fantasma: %v", err)
		}
		if removeErr != nil {
			t.Fatalf("remover fantasma: %v", removeErr)
		}
	}

	var startErr error
	err = redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CauseStart, func(r *rooms.Room) bool {
		if err := effects.StartGameEffect(ctx, redisFuncs.NewSink(e.rdb), r); err != nil {
			startErr = err
			return false
		}
		return true
	})
	if err != nil {
		t.Fatalf("iniciar partida: %v", err)
	}
	if startErr != nil {
		t.Fatalf("iniciar partida: %v", startErr)
	}

	winnerID := ids[0]
	invalidLeft := opts.invalid
	padLeft := opts.padInvalid
	disconnected := ""
	reconnectAt := 0
	invalidSeen := 0
	disconnectSeen := 0
	connectSeen := 0

	for step := 1; ; step++ {
		current, err := redisFuncs.LoadRoom(ctx, e.rdb, room.ID)
		if err != nil {
			t.Fatalf("carregar sala no passo %d: %v", step, err)
		}
		if current.GameOver {
			room = current
			break
		}
		if step > opts.maxSteps {
			t.Fatalf("partida não terminou em %d passos (evento %s, jogadores %d vivos)",
				opts.maxSteps, eventType(current), aliveCount(current))
		}

		// Desconecta um jogador que não está jogando agora; reconecta depois de
		// algumas entradas. O runner precisa aplicar o mesmo bitmap.
		if opts.disconnect && disconnected == "" && step == 8 {
			candidate := disconnectCandidate(current)
			if candidate != "" {
				if err := playerRedis.UnregisterPlayerFromRoom(ctx, e.rdb, candidate); err != nil {
					t.Fatalf("desregistrar %s: %v", candidate, err)
				}
				if err := redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CauseDisconnect, func(r *rooms.Room) bool {
					r.RemoveConnected(candidate)
					return true
				}); err != nil {
					t.Fatalf("desconectar %s: %v", candidate, err)
				}
				visual.step(step, "desconectou "+shortID(candidate), current)
				disconnected = candidate
				reconnectAt = step + 4
				disconnectSeen++
				continue
			}
		}
		if disconnected != "" && step >= reconnectAt {
			if err := playerRedis.RegisterPlayerInRoom(ctx, e.rdb, room.ID, disconnected); err != nil {
				t.Fatalf("registrar %s: %v", disconnected, err)
			}
			if err := redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CauseConnect, func(r *rooms.Room) bool {
				r.AddConnected(disconnected)
				return true
			}); err != nil {
				t.Fatalf("reconectar %s: %v", disconnected, err)
			}
			visual.step(step, "reconectou "+shortID(disconnected), current)
			disconnected = ""
			connectSeen++
			continue
		}

		if invalidLeft > 0 || padLeft > 0 {
			actor := outOfTurnActor(current)
			if actor != "" {
				if applied := play(ctx, e, room, actor, `{"type":"TAKE_COINS"}`); applied {
					t.Fatalf("jogada fora do turno foi aceita (ator %s, atual %s)", actor, current.CurrentPlayer)
				}
				visual.step(step, "jogada inválida (fora do turno)", current)
				if invalidLeft > 0 {
					invalidLeft--
				} else {
					padLeft--
				}
				invalidSeen++
				continue
			}
		}

		if current.GameEvent != nil && current.GameEvent.Type == structs.EventWaitingFirstAction &&
			current.CurrentPlayer != "" && current.IsConnected(current.CurrentPlayer) && isAlive(current, current.CurrentPlayer) {
			raw, ok := chooseAction(current, current.CurrentPlayer, winnerID, e.registry)
			if ok {
				if !play(ctx, e, room, current.CurrentPlayer, raw) {
					t.Fatalf("jogada principal recusada: %s (ator %s, evento %s)", raw, current.CurrentPlayer, eventType(current))
				}
				visual.step(step, "jogada "+shortAction(raw), current)
				continue
			}
		}

		if err := redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CauseTick, func(r *rooms.Room) bool {
			engine.Tick(ctx, redisFuncs.NewSink(e.rdb), e.registry, r)
			return true
		}); err != nil {
			t.Fatalf("tick: %v", err)
		}
		visual.step(step, "tick ("+eventType(current)+")", current)
	}

	finalRoom, err := redisFuncs.LoadRoom(ctx, e.rdb, room.ID)
	if err != nil {
		t.Fatalf("carregar sala final: %v", err)
	}
	if !finalRoom.GameOver {
		t.Fatal("sala terminou sem GameOver")
	}

	recorded, err := redisFuncs.ReadReplayEntries(ctx, e.rdb, room.MatchID)
	if err != nil {
		t.Fatalf("ler entradas de replay: %v", err)
	}
	if len(recorded) == 0 {
		t.Fatal("nenhuma entrada de replay gravada")
	}
	if recorded[0].Type != "start" {
		t.Fatalf("primeira entrada = %q, quer start", recorded[0].Type)
	}
	if err := services.FinalizeReplay(ctx, e.backend, e.rdb, finalRoom); err != nil {
		t.Fatalf("finalizar replay: %v", err)
	}
	envelope := e.takeEnvelope()
	if envelope == nil {
		t.Fatal("backend não recebeu o envelope do replay")
	}
	if envelope.RoomID != finalRoom.ID {
		t.Fatalf("envelope room_id = %q, quer %q", envelope.RoomID, finalRoom.ID)
	}

	raw, err := base64.StdEncoding.DecodeString(envelope.Data)
	if err != nil {
		t.Fatalf("envelope data não é base64: %v", err)
	}
	decoded, err := format.Decode(raw, nil)
	if err != nil {
		t.Fatalf("decodificar .hwr: %v", err)
	}
	checkHeader(t, decoded, finalRoom, envelope, opts.players)

	res := &matchResult{
		hwr:       raw,
		recorded:  recorded,
		finalRoom: finalRoom,
		envelope:  *envelope,
	}
	res.finalHash = replayAndCompare(t, decoded, recorded, finalRoom)
	visual.end(len(recorded), len(raw), res.finalHash)

	if opts.invalid > 0 && invalidSeen == 0 {
		t.Fatal("nenhuma jogada inválida foi gravada")
	}
	if opts.disconnect && (disconnectSeen == 0 || connectSeen == 0) {
		t.Fatalf("desconexão/reconexão não gravadas (disconnect=%d connect=%d)", disconnectSeen, connectSeen)
	}

	return res
}

// overrideRoom troca semente/matchId antes do início (para o golden) e refaz o
// baralho com o RNG novo, como em CreateNewRoom.
func overrideRoom(ctx context.Context, e *env, room *rooms.Room, opts matchOptions) error {
	var overrideErr error
	err := redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CauseJoin, func(r *rooms.Room) bool {
		if opts.seed != nil {
			r.Seed = *opts.seed
			r.Rng = rooms.NewRngState(*opts.seed)
			general, err := r.GetGeneralRules(e.registry)
			if err != nil {
				overrideErr = err
				return false
			}
			if general.CopysOfEachCard == nil {
				overrideErr = fmt.Errorf("regras sem copysOfEachCard")
				return false
			}
			deck := make([]structs.CardName, 0, *general.CopysOfEachCard*len(structs.AllCards))
			for i := 0; i < *general.CopysOfEachCard; i++ {
				deck = append(deck, structs.AllCards...)
			}
			r.Rand().Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
			r.Deck = deck
		}
		if opts.matchID != "" {
			r.MatchID = opts.matchID
		}
		r.Created = time.UnixMilli(1700000000000).UTC()
		return true
	})
	if err != nil {
		return err
	}
	return overrideErr
}

func addPlayer(ctx context.Context, e *env, room *rooms.Room, id, name string) error {
	if err := playerRedis.RegisterPlayerInRoom(ctx, e.rdb, room.ID, id); err != nil {
		return err
	}
	var addErr error
	err := redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CauseJoin, func(r *rooms.Room) bool {
		aplicado, err := redisFuncs.AddPlayerInRoom(e.registry, r, id, name)
		if err != nil {
			addErr = err
			return false
		}
		return aplicado
	})
	if err != nil {
		return err
	}
	return addErr
}

func play(ctx context.Context, e *env, room *rooms.Room, actor, raw string) bool {
	e.t.Helper()
	pp, err := structs.ParsePlayerPlay([]byte(raw), actor)
	if err != nil {
		e.t.Fatalf("payload de teste inválido (%s): %v", raw, err)
	}
	applied := false
	err = redisFuncs.WithRoom(ctx, e.rdb, room.ID, redisFuncs.CausePlay, func(r *rooms.Room) bool {
		applied = connection.OnMessage(ctx, e.rdb, e.registry, pp, r)
		return applied
	}, redisFuncs.ReplayPlayHint{Actor: actor, Payload: []byte(raw)})
	if err != nil {
		e.t.Fatalf("withRoom da jogada: %v", err)
	}
	return applied
}

func chooseAction(room *rooms.Room, actorID, winnerID string, registry *rules.Registry) (string, bool) {
	actor := room.Players[actorID]
	if actor == nil || !actor.Alive {
		return "", false
	}
	cost, err := room.GetBasicKillCardCost(registry)
	if err != nil {
		return "", false
	}
	if actor.Coins < cost {
		return `{"type":"TAKE_COINS"}`, true
	}

	if actorID == winnerID {
		if targetID, cardIndex, ok := pickTarget(room, func(id string) bool { return id != winnerID }); ok {
			return killPayload(targetID, cardIndex), true
		}
	} else {
		// Quem não é o vencedor nunca ataca o vencedor; se não há outro alvo,
		// coleta moedas para o vencedor ter a vez e eliminá-lo.
		if targetID, cardIndex, ok := pickTarget(room, func(id string) bool {
			return id != actorID && id != winnerID
		}); ok {
			return killPayload(targetID, cardIndex), true
		}
	}
	return `{"type":"TAKE_COINS"}`, true
}

func killPayload(playerID string, cardIndex int) string {
	return fmt.Sprintf(`{"type":"KILL_CARD","payload":{"targetPlayer":%q,"targetCardIndex":%d}}`, playerID, cardIndex)
}

// pickTarget percorre os ids ordenados e devolve a primeira carta viva de um
// jogador que satisfaça o filtro (nunca carta protegida/morta).
func pickTarget(room *rooms.Room, keep func(string) bool) (string, int, bool) {
	for _, id := range room.SortedPlayerIDs() {
		p := room.Players[id]
		if p == nil || !p.Alive || !keep(id) {
			continue
		}
		for _, card := range p.Cards {
			if !card.Dead && !card.Protected {
				return id, card.Index, true
			}
		}
	}
	return "", 0, false
}

func disconnectCandidate(room *rooms.Room) string {
	for _, id := range room.SortedPlayerIDs() {
		p := room.Players[id]
		if p == nil || !p.Alive || !room.IsConnected(id) || id == room.CurrentPlayer {
			continue
		}
		return id
	}
	return ""
}

func outOfTurnActor(room *rooms.Room) string {
	for _, id := range room.SortedPlayerIDs() {
		p := room.Players[id]
		if p == nil || !p.Alive {
			continue
		}
		if id != room.CurrentPlayer {
			return id
		}
	}
	return ""
}

func isAlive(room *rooms.Room, id string) bool {
	p := room.Players[id]
	return p != nil && p.Alive
}

func aliveCount(room *rooms.Room) int {
	n := 0
	for _, p := range room.Players {
		if p.Alive {
			n++
		}
	}
	return n
}

func eventType(room *rooms.Room) string {
	if room.GameEvent == nil {
		return "<nil>"
	}
	return string(room.GameEvent.Type)
}

// checkHeader confere o cabeçalho do .hwr contra a sala final e o envelope.
func checkHeader(t *testing.T, decoded *format.Replay, finalRoom *rooms.Room, envelope *uploadEnvelope, players int) {
	t.Helper()

	if decoded.EngineVersion != services.EngineVersion {
		t.Fatalf("engine version = %d, quer %d", decoded.EngineVersion, services.EngineVersion)
	}
	if decoded.Mode != string(finalRoom.Rules) {
		t.Fatalf("modo = %q, quer %q", decoded.Mode, finalRoom.Rules)
	}
	if decoded.StartedAtMs != finalRoom.StartTime.UnixMilli() {
		t.Fatalf("startedAt = %d, quer %d", decoded.StartedAtMs, finalRoom.StartTime.UnixMilli())
	}
	if decoded.Seed != [16]byte(finalRoom.Seed) {
		t.Fatal("semente do cabeçalho difere da sala")
	}
	if len(decoded.Players) != players {
		t.Fatalf("cabeçalho tem %d jogadores, quer %d", len(decoded.Players), players)
	}

	rank := replayIndexMap(finalRoom)
	roster := idsByMoveTime(finalRoom)
	for i, id := range roster {
		if got := decoded.Players[i].Index; got != rank[id] {
			t.Fatalf("jogador %d: índice = %d, quer %d", i, got, rank[id])
		}
	}
	if len(envelope.PlayerIDs) != len(roster) {
		t.Fatalf("envelope player_ids = %d, quer %d", len(envelope.PlayerIDs), len(roster))
	}
}

// replayAndCompare reproduz o .hwr e compara o hash de cada passo com o
// checkpoint que o servidor gravou naquela entrada. Também confere o fim.
func replayAndCompare(t *testing.T, decoded *format.Replay, recorded []redisFuncs.ReplayEntry, finalRoom *rooms.Room) string {
	t.Helper()

	r, err := runner.New(decoded, runner.Options{})
	if err != nil {
		t.Fatalf("criar runner: %v", err)
	}
	defer r.Close()

	if len(decoded.Entries) != len(recorded)+1 {
		t.Fatalf("entradas do .hwr = %d, quer %d (gravadas + fim)", len(decoded.Entries), len(recorded)+1)
	}

	for i := range recorded {
		if err := r.Step(); err != nil {
			t.Fatalf("passo %d/%d: %v", i+1, len(recorded), err)
		}
		got := r.Hash()
		want := hex.EncodeToString(recorded[i].CP[:])
		if !strings.HasPrefix(got, want) {
			t.Fatalf("estado divergiu no passo %d (%s): runner %s != servidor %s",
				i+1, recorded[i].Type, got[:16], want)
		}
	}

	if err := r.Step(); err != nil {
		t.Fatalf("passo final (fim): %v", err)
	}
	if err := r.Step(); err != runner.ErrDone {
		t.Fatalf("depois do fim esperava ErrDone, veio %v", err)
	}

	if r.State() == nil || !r.GameOver() {
		t.Fatal("runner não terminou em GameOver")
	}
	wantWinner := ""
	if finalRoom.WinnerID != "" {
		wantWinner = replayIndexString(finalRoom, finalRoom.WinnerID)
	}
	if got := r.WinnerID(); got != wantWinner {
		t.Fatalf("vencedor no runner = %q, quer %q", got, wantWinner)
	}
	return r.Hash()
}

// replayIndexMap devolve o índice canônico (ordem lexicográfica dos UUIDs) de
// cada jogador, a mesma regra do recorder.
func replayIndexMap(room *rooms.Room) map[string]uint8 {
	ids := make([]string, 0, len(room.Players))
	for id := range room.Players {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	m := make(map[string]uint8, len(ids))
	for i, id := range ids {
		m[id] = uint8(i)
	}
	return m
}

func replayIndexString(room *rooms.Room, id string) string {
	idx, ok := replayIndexMap(room)[id]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%d", idx)
}

func idsByMoveTime(room *rooms.Room) []string {
	ids := make([]string, 0, len(room.Players))
	for id := range room.Players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := room.Players[ids[i]], room.Players[ids[j]]
		if a.MoveTime != b.MoveTime {
			return a.MoveTime < b.MoveTime
		}
		return ids[i] < ids[j]
	})
	return ids
}

// TestReplayE2E roda 200 partidas (5 com -short) com jogadas inválidas,
// desconexões e buracos de MoveTime. Exige estado igual em TODOS os passos e
// arquivo ≤ 4 KB (mediana e máximo).
func TestReplayE2E(t *testing.T) {
	games := 200
	if testing.Short() {
		games = 5
	}

	e := newEnv(t)
	sizes := make([]int, 0, games)

	for i := 0; i < games; i++ {
		i := i
		t.Run(fmt.Sprintf("partida_%03d", i), func(t *testing.T) {
			opts := matchOptions{
				label:          fmt.Sprintf("partida %03d/%03d", i+1, games),
				players:        3,
				invalid:        2,
				disconnect:     true,
				gapBeforeStart: i%4 == 0,
			}
			if i >= 5 {
				opts.players = 2
			}
			if i == 0 {
				// Partida longa: centenas de jogadas recusadas (cada uma vira
				// entrada no replay) sem mudar o estado.
				opts.padInvalid = 300
				opts.maxSteps = 4000
			}
			res := playMatch(t, e, opts)
			sizes = append(sizes, len(res.hwr))
		})
	}

	if len(sizes) != games {
		t.Fatalf("rodaram %d partidas, quer %d", len(sizes), games)
	}
	sort.Ints(sizes)
	median := sizes[len(sizes)/2]
	maxSize := sizes[len(sizes)-1]
	if median > 4096 {
		t.Fatalf("tamanho mediano = %d bytes (alvo ≤ 4096)", median)
	}
	if maxSize > 4096 {
		t.Fatalf("tamanho máximo = %d bytes (alvo ≤ 4096)", maxSize)
	}
	t.Logf("tamanhos: mediano=%d B, máximo=%d B, mínimo=%d B", median, maxSize, sizes[0])
	visualSummary(games, sizes[0], median, maxSize)
}

// TestGoldenV1 reproduz o golden congelado da v1. Rode com UPDATE_GOLDEN=1
// para regravá-lo a partir do fluxo real (recorder + runner).
func TestGoldenV1(t *testing.T) {
	silenceEngineLogs()
	path := filepath.Join("..", "engines", "v1", "testdata", goldenFileName)
	hashPath := strings.TrimSuffix(path, ".hwr") + ".sha256"

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		e := newEnv(t)
		seed := fixedGoldenSeed()
		res := playMatch(t, e, matchOptions{
			label:          "golden v1",
			players:        3,
			names:          []string{"ana", "bia", "cris"},
			seed:           &seed,
			matchID:        goldenMatchID,
			gapBeforeStart: true,
			invalid:        2,
			disconnect:     true,
			maxSteps:       4000,
		})
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("criar testdata: %v", err)
		}
		if err := os.WriteFile(path, res.hwr, 0o644); err != nil {
			t.Fatalf("gravar golden: %v", err)
		}
		if err := os.WriteFile(hashPath, []byte(res.finalHash+"\n"), 0o644); err != nil {
			t.Fatalf("gravar hash do golden: %v", err)
		}
		t.Logf("golden gravado: %d bytes, hash final %s", len(res.hwr), res.finalHash)
		return
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ler golden (rode UPDATE_GOLDEN=1 para gerar): %v", err)
	}
	wantBytes, err := os.ReadFile(hashPath)
	if err != nil {
		t.Fatalf("ler hash do golden: %v", err)
	}
	want := strings.TrimSpace(string(wantBytes))

	decoded, err := format.Decode(raw, nil)
	if err != nil {
		t.Fatalf("decodificar golden: %v", err)
	}
	r, err := runner.New(decoded, runner.Options{})
	if err != nil {
		t.Fatalf("criar runner: %v", err)
	}
	defer r.Close()

	for r.Position() < r.Total() {
		if err := r.Step(); err != nil {
			t.Fatalf("golden, passo %d: %v", r.Position()+1, err)
		}
	}
	if got := r.Hash(); got != want {
		t.Fatalf("hash final do golden = %s, quer %s", got, want)
	}
	if r.State() == nil || !r.GameOver() {
		t.Fatal("golden não termina em GameOver")
	}
	t.Logf("golden: %d bytes, %d entradas, hash %s", len(raw), len(decoded.Entries), want)
}
