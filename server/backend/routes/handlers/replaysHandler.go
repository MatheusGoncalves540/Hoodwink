package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
	"github.com/MatheusGoncalves540/Hoodwink-replay/versions"
	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// maxReplayBody é o teto do POST /internal/replays (mesmo teto do .hwr).
	maxReplayBody = 256 * 1024
	// defaultReplayMatches é o limit padrão do GET /me/matches.
	defaultReplayMatches = 30
	// maxReplayMatches é o teto de limit do GET /me/matches.
	maxReplayMatches = 100
	// replayRetentionWindow é a janela do histórico (últimos 30 dias).
	replayRetentionWindow = 30 * 24 * time.Hour
	// replayWriteTimeout limita cada consulta/escrita de replay.
	replayWriteTimeout = 5 * time.Second
)

// replayStore junta o que os handlers de replay precisam do Postgres.
type replayStore interface {
	// Insert grava com ON CONFLICT (match_id) DO NOTHING: repetido é no-op.
	Insert(ctx context.Context, r *models.MatchReplay) error
	// ListByPlayer devolve as partidas do jogador nos últimos 30 dias.
	ListByPlayer(ctx context.Context, playerID string, limit int, since time.Time) ([]models.MatchReplay, error)
	// FindByMatchAndPlayer devolve o replay só se match_id existe E o jogador
	// participou (player_ids contém playerID); senão ErrRecordNotFound.
	FindByMatchAndPlayer(ctx context.Context, matchID, playerID string) (*models.MatchReplay, error)
}

// gormReplayStore é a implementação real com gorm/Postgres.
type gormReplayStore struct {
	db *gorm.DB
}

const insertReplaySQL = `
INSERT INTO match_replays (match_id, room_id, engine_version, game_version, started_at, ended_at, winner_id, player_ids, data)
VALUES (?, ?, ?, ?, ?, ?, ?, ?::uuid[], ?)
ON CONFLICT (match_id) DO NOTHING`

const listReplaysSQL = `
SELECT match_id, room_id, engine_version, game_version, started_at, ended_at, winner_id, player_ids
FROM match_replays
WHERE player_ids @> ARRAY[?]::uuid[] AND ended_at >= ?
ORDER BY ended_at DESC
LIMIT ?`

const findReplaySQL = `
SELECT match_id, room_id, engine_version, game_version, started_at, ended_at, winner_id, player_ids, data
FROM match_replays
WHERE match_id = ? AND player_ids @> ARRAY[?]::uuid[]`

func (s gormReplayStore) Insert(ctx context.Context, r *models.MatchReplay) error {
	return s.db.WithContext(ctx).Exec(insertReplaySQL,
		r.MatchID, r.RoomID, r.EngineVersion, r.GameVersion, r.StartedAt, r.EndedAt, r.WinnerID,
		uuidArrayLiteral(r.PlayerIDs), r.Data,
	).Error
}

func (s gormReplayStore) ListByPlayer(ctx context.Context, playerID string, limit int, since time.Time) ([]models.MatchReplay, error) {
	var rows []models.MatchReplay
	if err := s.db.WithContext(ctx).Raw(listReplaysSQL, playerID, since, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s gormReplayStore) FindByMatchAndPlayer(ctx context.Context, matchID, playerID string) (*models.MatchReplay, error) {
	var rows []models.MatchReplay
	if err := s.db.WithContext(ctx).Raw(findReplaySQL, matchID, playerID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &rows[0], nil
}

// uuidArrayLiteral monta o literal de array do Postgres `{a,b,c}`. UUIDs não
// contêm vírgula nem chaves, então a concatenação é segura.
func uuidArrayLiteral(ids []string) string {
	return "{" + strings.Join(ids, ",") + "}"
}

// gameVersionRE aceita semver e o identificador "dev" (build sem ldflags).
var gameVersionRE = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z.+-]{0,31}$`)

// sanitizeGameVersion normaliza a versão do jogo enviada pelo gameServer.
// Valor vazio, grande demais ou fora do formato vira "unknown".
func sanitizeGameVersion(v string) string {
	v = strings.TrimSpace(v)
	if !gameVersionRE.MatchString(v) {
		return "unknown"
	}
	return v
}

// replayUpload é o corpo do POST /internal/replays. data é o .hwr (base64).
// room_id e player_ids vêm do gameServer: o cabeçalho .hwr só tem matchId,
// versão, início/fim e jogadores (índice+nome); os UUIDs dos jogadores chegam
// aqui e são conferidos contra a contagem do cabeçalho. winner/vencedor no corpo
// é ignorado (o vencedor é derivado da última entrada EntryEnd do cabeçalho).
type replayUpload struct {
	Data        string   `json:"data"`
	RoomID      string   `json:"room_id"`
	PlayerIDs   []string `json:"player_ids"`
	GameVersion string   `json:"game_version"`
}

// matchDTO é o item do histórico devolvido pelo GET /me/matches (sem o blob).
type matchDTO struct {
	MatchID       string    `json:"match_id"`
	RoomID        string    `json:"room_id,omitempty"`
	EngineVersion int       `json:"engine_version"`
	GameVersion   string    `json:"game_version"`
	StartedAt     time.Time `json:"started_at"`
	EndedAt       time.Time `json:"ended_at"`
	WinnerID      *string   `json:"winner_id,omitempty"`
	PlayerIDs     []string  `json:"player_ids"`
}

func toMatchDTO(m models.MatchReplay) matchDTO {
	return matchDTO{
		MatchID:       m.MatchID,
		RoomID:        m.RoomID,
		EngineVersion: m.EngineVersion,
		GameVersion:   m.GameVersion,
		StartedAt:     m.StartedAt,
		EndedAt:       m.EndedAt,
		WinnerID:      m.WinnerID,
		PlayerIDs:     []string(m.PlayerIDs),
	}
}

// InternalReplaysHandler recebe o replay do gameServer (plano 05, T9).
func (h *Handler) InternalReplaysHandler(w http.ResponseWriter, r *http.Request) {
	writeReplayUpload(w, r, h.ReplayStore, h.ReplayService.Signer())
}

// writeReplayUpload decodifica, valida, assina e grava. Ordem: decode → validação
// (versão, fim, jogadores) → assinatura → insert (ON CONFLICT DO NOTHING).
func writeReplayUpload(w http.ResponseWriter, r *http.Request, store replayStore, signer format.Signer) {
	r.Body = http.MaxBytesReader(w, r.Body, maxReplayBody)

	var body replayUpload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			utils.SendError(w, "corpo muito grande", http.StatusRequestEntityTooLarge)
			return
		}
		utils.SendError(w, "payload inválido", http.StatusBadRequest)
		return
	}

	raw, err := base64.StdEncoding.DecodeString(body.Data)
	if err != nil || len(raw) == 0 {
		utils.SendError(w, "data inválida", http.StatusBadRequest)
		return
	}

	replay, err := format.Decode(raw, nil)
	if err != nil {
		utils.SendError(w, "replay inválido", http.StatusBadRequest)
		return
	}

	if _, ok := versions.ByNumber(replay.EngineVersion); !ok {
		utils.SendError(w, "versão do motor não suportada", http.StatusBadRequest)
		return
	}

	winnerIdx, ok := replayWinner(replay)
	if !ok {
		utils.SendError(w, "replay sem fim", http.StatusConflict)
		return
	}

	// player_ids do corpo é a fonte dos UUIDs (o cabeçalho só tem índice+nome);
	// a contagem e a ordem vêm do cabeçalho. winner_id é derivado do cabeçalho.
	if len(body.PlayerIDs) != len(replay.Players) {
		utils.SendError(w, "player_ids não confere com o cabeçalho", http.StatusBadRequest)
		return
	}
	playerIDs := make([]string, len(body.PlayerIDs))
	for i, id := range body.PlayerIDs {
		if _, err := uuid.Parse(id); err != nil {
			utils.SendError(w, "player_ids inválido", http.StatusBadRequest)
			return
		}
		playerIDs[i] = id
	}

	matchID := uuid.UUID(replay.MatchID)
	started := time.UnixMilli(replay.StartedAtMs).UTC()
	ended := started.Add(time.Duration(replayTotalDeltaMs(replay)) * time.Millisecond)

	winner := playerIDs[winnerIdx]
	record := &models.MatchReplay{
		MatchID:       matchID.String(),
		RoomID:        body.RoomID,
		EngineVersion: replay.EngineVersion,
		GameVersion:   sanitizeGameVersion(body.GameVersion),
		StartedAt:     started,
		EndedAt:       ended,
		WinnerID:      &winner,
		PlayerIDs:     playerIDs,
	}

	// Assina antes de gravar; o data persistido é o arquivo assinado.
	signed, err := format.Encode(replay, signer)
	if err != nil {
		utils.SendError(w, "replay inválido", http.StatusBadRequest)
		return
	}
	record.Data = signed

	ctx, cancel := context.WithTimeout(r.Context(), replayWriteTimeout)
	defer cancel()
	if err := store.Insert(ctx, record); err != nil {
		utils.SendError(w, "erro ao gravar replay", http.StatusInternalServerError)
		return
	}

	// Novo e repetido (ON CONFLICT DO NOTHING) devolvem 200; nunca sobrescreve.
	utils.SendSuccess(w, map[string]string{"match_id": record.MatchID}, "")
}

// MeMatchesHandler devolve o histórico do usuário do token (últimos 30 dias).
func (h *Handler) MeMatchesHandler(w http.ResponseWriter, r *http.Request) {
	writeMeMatches(w, r, h.ReplayStore)
}

func writeMeMatches(w http.ResponseWriter, r *http.Request, store replayStore) {
	user, ok := r.Context().Value(UserContextKey).(structures.UserClaims)
	if !ok || user.Id == "" {
		utils.SendError(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	limit := defaultReplayMatches
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			utils.SendError(w, "limit inválido", http.StatusBadRequest)
			return
		}
		if n > maxReplayMatches {
			n = maxReplayMatches
		}
		limit = n
	}

	// userId/outros parâmetros da query são ignorados: a identidade vem do token.
	since := time.Now().UTC().Add(-replayRetentionWindow)

	ctx, cancel := context.WithTimeout(r.Context(), replayWriteTimeout)
	defer cancel()
	matches, err := store.ListByPlayer(ctx, user.Id, limit, since)
	if err != nil {
		utils.SendError(w, "erro ao listar partidas", http.StatusInternalServerError)
		return
	}

	dtos := make([]matchDTO, 0, len(matches))
	for _, m := range matches {
		dtos = append(dtos, toMatchDTO(m))
	}
	utils.SendSuccess(w, dtos, "")
}

// ReplayHandler baixa o .hwr assinado de uma partida do usuário do token.
func (h *Handler) ReplayHandler(w http.ResponseWriter, r *http.Request) {
	writeReplay(w, r, h.ReplayStore)
}

func writeReplay(w http.ResponseWriter, r *http.Request, store replayStore) {
	user, ok := r.Context().Value(UserContextKey).(structures.UserClaims)
	if !ok || user.Id == "" {
		utils.SendError(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	matchID := chi.URLParam(r, "matchId")
	if _, err := uuid.Parse(matchID); err != nil {
		utils.SendError(w, "matchId inválido", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), replayWriteTimeout)
	defer cancel()
	record, err := store.FindByMatchAndPlayer(ctx, matchID, user.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 404 igual para inexistente e para quem não jogou.
			utils.SendError(w, "replay não encontrado", http.StatusNotFound)
			return
		}
		utils.SendError(w, "erro ao buscar replay", http.StatusInternalServerError)
		return
	}

	name := replayFilename(record)

	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(record.Data)
}

// replayFilename monta hoodwink-<data>-<matchId8>.hwr no servidor.
func replayFilename(m *models.MatchReplay) string {
	hex := strings.ReplaceAll(m.MatchID, "-", "")
	if len(hex) > 8 {
		hex = hex[:8]
	}
	return fmt.Sprintf("hoodwink-%s-%s.hwr", m.EndedAt.UTC().Format("2006-01-02"), hex)
}

// replayWinner devolve o índice do vencedor a partir da última entrada (EntryEnd).
func replayWinner(r *format.Replay) (int, bool) {
	if len(r.Entries) == 0 {
		return 0, false
	}
	last := r.Entries[len(r.Entries)-1]
	if last.Type != format.EntryEnd || last.Player == format.NoPlayer || int(last.Player) >= len(r.Players) {
		return 0, false
	}
	return int(last.Player), true
}

// replayTotalDeltaMs soma os deltas para derivar ended_at a partir de started_at.
func replayTotalDeltaMs(r *format.Replay) int64 {
	var total int64
	for _, e := range r.Entries {
		total += e.DeltaMs
	}
	return total
}
