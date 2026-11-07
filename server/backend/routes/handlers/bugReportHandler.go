package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	maxBugReportBody     = 16 * 1024
	maxBugContextBytes   = 8 * 1024
	maxBugDescription    = 2000
	minLobbyDescription  = 10
	maxTurn              = 10000
	maxPartidaReports24h = 50
	partidaWindow        = 24 * time.Hour
	matchFutureTolerance = 5 * time.Minute

	bugUserRateLimit = 10
	bugIPRateLimit   = 5
	bugRateWindow    = 10 * time.Minute
	bugRateMaxKeys   = 10000
)

// bugCategories é a lista fechada de categorias (espelha BUG_CATEGORIES do front).
var bugCategories = []string{"moedas", "carta", "travou", "regra", "tela", "outro"}

func validBugCategory(c string) bool {
	for _, v := range bugCategories {
		if c == v {
			return true
		}
	}
	return false
}

// bugReportStore junta o que os handlers de relato precisam do Postgres.
type bugReportStore interface {
	bugFlagStore
	bugCreateStore
	bugPatchStore
	bugAdminStore
}

// bugFlagStore cobre o POST /bugReports/flag.
type bugFlagStore interface {
	CountPartida24h(ctx context.Context, userID string) (int64, error)
	UpsertTurn(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (id int64, turns []int32, created bool, err error)
}

// bugCreateStore cobre o POST /bugReports (Lobby).
type bugCreateStore interface {
	Create(ctx context.Context, report *models.BugReport) error
}

// bugPatchStore cobre o PATCH /bugReports/{id}.
type bugPatchStore interface {
	Find(ctx context.Context, id int64) (*models.BugReport, error)
	Update(ctx context.Context, id int64, userID string, cat, desc *string, context json.RawMessage) (*models.BugReport, error)
}

// bugAdminStore cobre o GET/PATCH /admin/bugReports.
type bugAdminStore interface {
	List(ctx context.Context, f bugListFilter) (items []models.BugReport, total, novos int64, err error)
	Find(ctx context.Context, id int64) (*models.BugReport, error)
	UpdateStatus(ctx context.Context, id int64, status string) (*models.BugReport, error)
}

// gormBugReportStore é a implementação real com gorm/Postgres.
type gormBugReportStore struct {
	db *gorm.DB
}

// --- flag ---

func (s gormBugReportStore) CountPartida24h(ctx context.Context, userID string) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).
		Model(&models.BugReport{}).
		Where("user_id = ? AND source = ? AND created_at >= ?", userID, "partida", time.Now().UTC().Add(-24*time.Hour)).
		Count(&n).Error
	return n, err
}

// upsertTurnSQL acrescenta o turno de forma atômica: cria o relato se não existir
// ou soma o turno ao já existente (idempotente; não passa de 10 turnos).
const upsertTurnSQL = `
INSERT INTO bug_reports (user_id, username, source, match_id, turns, ip, user_agent, status)
VALUES (?, ?, 'partida', ?, ARRAY[?]::int[], ?, ?, 'novo')
ON CONFLICT (user_id, match_id) WHERE user_id IS NOT NULL AND match_id IS NOT NULL
DO UPDATE SET
	turns = CASE
		WHEN bug_reports.turns @> ARRAY[?]::int[] THEN bug_reports.turns
		WHEN array_length(bug_reports.turns, 1) >= 10 THEN bug_reports.turns
		ELSE array_append(bug_reports.turns, ?)
	END,
	updated_at = now()
RETURNING id, turns, (xmax = 0) AS created`

func (s gormBugReportStore) UpsertTurn(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
	var id int64
	var turns models.Int32Array
	var created bool
	err := s.db.WithContext(ctx).Raw(upsertTurnSQL, userID, username, matchID, turn, ip, ua, turn, turn).
		Row().
		Scan(&id, &turns, &created)
	return id, []int32(turns), created, err
}

// --- lobby ---

// insertLobbyBugSQL grava o relato do Lobby em SQL cru: turns fica no default
// (`{}`) porque nenhuma escrita de array pode passar pelo driver.Valuer, que
// devolve text e o Postgres recusa em int[]/uuid[]. RETURNING repõe id e datas.
const insertLobbyBugSQL = `
INSERT INTO bug_reports (user_id, username, source, match_id, category, description, context, ip, user_agent, status)
VALUES (?, ?, ?, ?, ?, ?, ?::jsonb, ?, ?, ?)
RETURNING id, created_at, updated_at`

func (s gormBugReportStore) Create(ctx context.Context, report *models.BugReport) error {
	return s.db.WithContext(ctx).Raw(insertLobbyBugSQL,
		report.UserID, report.Username, report.Source, report.MatchID,
		report.Category, report.Description, report.Context, report.IP, report.UserAgent, report.Status,
	).Row().Scan(&report.ID, &report.CreatedAt, &report.UpdatedAt)
}

// --- patch / admin ---

func (s gormBugReportStore) Find(ctx context.Context, id int64) (*models.BugReport, error) {
	var r models.BugReport
	if err := s.db.WithContext(ctx).First(&r, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &r, nil
}

func (s gormBugReportStore) Update(ctx context.Context, id int64, userID string, cat, desc *string, context json.RawMessage) (*models.BugReport, error) {
	updates := map[string]any{}
	if cat != nil {
		updates["category"] = *cat
	}
	if desc != nil {
		updates["description"] = *desc
	}
	if context != nil {
		updates["context"] = context
	}
	if len(updates) > 0 {
		updates["updated_at"] = time.Now()
		res := s.db.WithContext(ctx).
			Model(&models.BugReport{}).
			Where("id = ? AND user_id = ? AND source = ?", id, userID, "partida").
			Updates(updates)
		if res.Error != nil {
			return nil, res.Error
		}
	}
	return s.Find(ctx, id)
}

func (s gormBugReportStore) List(ctx context.Context, f bugListFilter) ([]models.BugReport, int64, int64, error) {
	base := s.db.WithContext(ctx).Model(&models.BugReport{})
	if f.status != "" {
		base = base.Where("status = ?", f.status)
	}
	if f.source != "" {
		base = base.Where("source = ?", f.source)
	}
	if f.matchID != "" {
		base = base.Where("match_id = ?", f.matchID)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// novos: contagem de status='novo' respeitando source/matchId (o filtro de
	// status não deve esconder o contador da aba).
	novosBase := s.db.WithContext(ctx).Model(&models.BugReport{}).Where("status = ?", "novo")
	if f.source != "" {
		novosBase = novosBase.Where("source = ?", f.source)
	}
	if f.matchID != "" {
		novosBase = novosBase.Where("match_id = ?", f.matchID)
	}
	var novos int64
	if err := novosBase.Count(&novos).Error; err != nil {
		return nil, 0, 0, err
	}

	items := make([]models.BugReport, 0)
	if err := base.Order("created_at DESC").Limit(f.limit).Offset((f.page - 1) * f.limit).Find(&items).Error; err != nil {
		return nil, 0, 0, err
	}

	return items, total, novos, nil
}

func (s gormBugReportStore) UpdateStatus(ctx context.Context, id int64, status string) (*models.BugReport, error) {
	res := s.db.WithContext(ctx).
		Model(&models.BugReport{}).
		Where("id = ?", id).
		Updates(map[string]any{"status": status, "updated_at": time.Now()})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return s.Find(ctx, id)
}

// --- DTO e resposta ---

type bugReportDTO struct {
	ID          int64           `json:"id"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	UserID      *string         `json:"user_id"`
	Username    *string         `json:"username"`
	Source      string          `json:"source"`
	MatchID     *string         `json:"match_id"`
	Turns       []int32         `json:"turns"`
	Category    *string         `json:"category"`
	Description *string         `json:"description"`
	Context     json.RawMessage `json:"context"`
	IP          *string         `json:"ip"`
	UserAgent   *string         `json:"user_agent"`
	Status      string          `json:"status"`
}

func toBugReportDTO(m models.BugReport) bugReportDTO {
	turns := []int32(m.Turns)
	if turns == nil {
		turns = []int32{}
	}
	return bugReportDTO{
		ID:          m.ID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		UserID:      m.UserID,
		Username:    m.Username,
		Source:      m.Source,
		MatchID:     m.MatchID,
		Turns:       turns,
		Category:    m.Category,
		Description: m.Description,
		Context:     m.Context,
		IP:          m.IP,
		UserAgent:   m.UserAgent,
		Status:      m.Status,
	}
}

func writeBugJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// --- helpers ---

// decodeBugBody lê o corpo com MaxBytesReader de 16 KB → 413 e decodifica JSON.
func decodeBugBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBugReportBody)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			utils.SendError(w, "Corpo muito grande", http.StatusRequestEntityTooLarge)
			return false
		}
		utils.SendError(w, "payload inválido", http.StatusBadRequest)
		return false
	}
	return true
}

// sanitizeBugContext valida e mascara o context (até 8 KB). nil/vazio/null → nil.
func sanitizeBugContext(raw json.RawMessage) (json.RawMessage, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	if !json.Valid(trimmed) {
		return nil, errors.New("contexto não é JSON")
	}
	if len(trimmed) > maxBugContextBytes {
		return nil, errors.New("contexto muito grande")
	}
	return utils.MaskContext(trimmed, maxBugContextBytes), nil
}

func containsTurn(turns []int32, turn int32) bool {
	for _, t := range turns {
		if t == turn {
			return true
		}
	}
	return false
}

func validMatchID(s string, now time.Time) bool {
	u, err := uuid.Parse(s)
	if err != nil || u.Version() != uuid.Version(7) {
		return false
	}
	t := uuidV7Time(u)
	return !t.Before(now.Add(-partidaWindow)) && !t.After(now.Add(matchFutureTolerance))
}

// uuidV7Time extrai o timestamp (milissegundos) dos primeiros 48 bits de um v7.
func uuidV7Time(u uuid.UUID) time.Time {
	b := u[:]
	ms := int64(b[0])<<40 | int64(b[1])<<32 | int64(b[2])<<24 | int64(b[3])<<16 | int64(b[4])<<8 | int64(b[5])
	return time.UnixMilli(ms).UTC()
}

// --- handlers ---

// optionalUser valida um Bearer opcional para o Lobby; ausente/inválido/vencido
// devolve false (anônimo), nunca um erro.
func (h *Handler) optionalUser(r *http.Request) (structures.UserClaims, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return structures.UserClaims{}, false
	}
	user, err := h.JWTService.ValidateJWT(strings.TrimPrefix(auth, "Bearer "))
	if err != nil || user.Id == "" {
		return structures.UserClaims{}, false
	}
	return user, true
}

// allowBugRate aplica o limite em memória: logado por usuário, anônimo por IP.
func (h *Handler) allowBugRate(w http.ResponseWriter, r *http.Request, userID string, logged bool) bool {
	limiter := h.BugIPLimiter
	key := ipRateKey(utils.ClientIP(r))
	if logged {
		limiter = h.BugUserLimiter
		key = "u:" + userID
	}
	if limiter == nil || limiter.Allow(key) {
		return true
	}
	utils.SendError(w, "Muitas tentativas. Tente novamente em alguns minutos.", http.StatusTooManyRequests)
	return false
}

// BugReportFlagHandler marca um turno na partida (login obrigatório).
func (h *Handler) BugReportFlagHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(structures.UserClaims)
	if !ok || user.Id == "" {
		utils.SendError(w, "Token inválido", http.StatusUnauthorized)
		return
	}
	if !h.allowBugRate(w, r, user.Id, true) {
		return
	}
	flagBugReport(w, r, user, h.BugReportStore, time.Now())
}

func flagBugReport(w http.ResponseWriter, r *http.Request, user structures.UserClaims, store bugFlagStore, now time.Time) {
	var req struct {
		MatchID string `json:"matchId"`
		Turn    int32  `json:"turn"`
	}
	if !decodeBugBody(w, r, &req) {
		return
	}

	if !validMatchID(req.MatchID, now) {
		utils.SendError(w, "matchId inválido", http.StatusBadRequest)
		return
	}
	if req.Turn < 0 || req.Turn > maxTurn {
		utils.SendError(w, "turno inválido", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	n, err := store.CountPartida24h(ctx, user.Id)
	if err != nil {
		utils.SendError(w, "erro ao registrar relato", http.StatusInternalServerError)
		return
	}
	if n >= maxPartidaReports24h {
		utils.SendError(w, "limite de relatos por partida atingido", http.StatusTooManyRequests)
		return
	}

	id, turns, created, err := store.UpsertTurn(ctx, user.Id, user.Username, req.MatchID, utils.ClientIP(r), utils.UserAgent(r), req.Turn)
	if err != nil {
		utils.SendError(w, "erro ao registrar relato", http.StatusInternalServerError)
		return
	}
	if !containsTurn(turns, req.Turn) {
		utils.SendError(w, "máximo de 10 turnos por partida", http.StatusBadRequest)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeBugJSON(w, status, map[string]any{
		"id":       id,
		"match_id": req.MatchID,
		"turns":    turns,
	})
}

// BugReportLobbyHandler cria um relato do Lobby (auth opcional).
func (h *Handler) BugReportLobbyHandler(w http.ResponseWriter, r *http.Request) {
	user, logged := h.optionalUser(r)
	if !h.allowBugRate(w, r, user.Id, logged) {
		return
	}
	createLobbyBug(w, r, user, logged, h.BugReportStore)
}

type lobbyBugRequest struct {
	Category    string          `json:"category"`
	Description string          `json:"description"`
	Context     json.RawMessage `json:"context"`
}

func createLobbyBug(w http.ResponseWriter, r *http.Request, user structures.UserClaims, logged bool, store bugCreateStore) {
	var req lobbyBugRequest
	if !decodeBugBody(w, r, &req) {
		return
	}

	if !validBugCategory(req.Category) {
		utils.SendError(w, "categoria inválida", http.StatusBadRequest)
		return
	}
	if n := utf8.RuneCountInString(req.Description); n < minLobbyDescription || n > maxBugDescription {
		utils.SendError(w, "descrição deve ter entre 10 e 2000 caracteres", http.StatusBadRequest)
		return
	}

	ctxJSON, err := sanitizeBugContext(req.Context)
	if err != nil {
		utils.SendError(w, "contexto inválido", http.StatusBadRequest)
		return
	}

	var userID, username *string
	if logged {
		userID = &user.Id
		username = &user.Username
	}

	ip := utils.ClientIP(r)
	ua := utils.UserAgent(r)
	source := "lobby"
	report := &models.BugReport{
		UserID:      userID,
		Username:    username,
		Source:      source,
		Category:    &req.Category,
		Description: &req.Description,
		Context:     ctxJSON,
		IP:          &ip,
		UserAgent:   &ua,
		Status:      "novo",
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := store.Create(ctx, report); err != nil {
		utils.SendError(w, "erro ao registrar relato", http.StatusInternalServerError)
		return
	}

	writeBugJSON(w, http.StatusCreated, toBugReportDTO(*report))
}

// BugReportPatchHandler detalha um relato de partida (só o dono).
func (h *Handler) BugReportPatchHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(structures.UserClaims)
	if !ok || user.Id == "" {
		utils.SendError(w, "Token inválido", http.StatusUnauthorized)
		return
	}
	if !h.allowBugRate(w, r, user.Id, true) {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		utils.SendError(w, "id inválido", http.StatusBadRequest)
		return
	}
	patchBugReport(w, r, id, user.Id, h.BugReportStore)
}

func patchBugReport(w http.ResponseWriter, r *http.Request, id int64, userID string, store bugPatchStore) {
	var req struct {
		Category    string          `json:"category"`
		Description *string         `json:"description"`
		Context     json.RawMessage `json:"context"`
	}
	if !decodeBugBody(w, r, &req) {
		return
	}

	if req.Category != "" && !validBugCategory(req.Category) {
		utils.SendError(w, "categoria inválida", http.StatusBadRequest)
		return
	}
	var cat *string
	if req.Category != "" {
		cat = &req.Category
	}
	var desc *string
	if req.Description != nil {
		if utf8.RuneCountInString(*req.Description) > maxBugDescription {
			utils.SendError(w, "descrição muito longa", http.StatusBadRequest)
			return
		}
		desc = req.Description
	}
	ctxJSON, err := sanitizeBugContext(req.Context)
	if err != nil {
		utils.SendError(w, "contexto inválido", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	report, err := store.Find(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendError(w, "relato não encontrado", http.StatusNotFound)
			return
		}
		utils.SendError(w, "erro ao buscar relato", http.StatusInternalServerError)
		return
	}

	if report.Source != "partida" || report.UserID == nil || *report.UserID != userID {
		utils.SendError(w, "Acesso negado", http.StatusForbidden)
		return
	}

	updated, err := store.Update(ctx, id, userID, cat, desc, ctxJSON)
	if err != nil {
		utils.SendError(w, "erro ao atualizar relato", http.StatusInternalServerError)
		return
	}

	writeBugJSON(w, http.StatusOK, toBugReportDTO(*updated))
}
