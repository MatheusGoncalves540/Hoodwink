package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

const (
	// maxIngestBody limita o corpo do POST /internal/requestLogs (plano 04).
	maxIngestBody = 2 * 1024 * 1024
	// maxIngestLines limita o número de linhas por lote (plano 04).
	maxIngestLines = 500

	truncPath      = 512
	truncUserAgent = 256
	truncUsername  = 64
	truncBody      = 4096
	truncError     = 1024
)

// requestLogInserter é o que o ingest precisa do RequestLogService.
type requestLogInserter interface {
	InsertBatch(ctx context.Context, entries []*models.RequestLog) error
}

// ingestRequestLog é o formato de cada linha enviada pelo gameServer. O service
// é ignorado e forçado para gameServer no servidor.
type ingestRequestLog struct {
	At         time.Time `json:"at"`
	Service    string    `json:"service"`
	Method     *string   `json:"method,omitempty"`
	Path       *string   `json:"path,omitempty"`
	Route      *string   `json:"route,omitempty"`
	Status     *int      `json:"status,omitempty"`
	DurationMs *int      `json:"duration_ms,omitempty"`
	IP         *string   `json:"ip,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	UserID     *string   `json:"user_id,omitempty"`
	Username   *string   `json:"username,omitempty"`
	RequestID  *string   `json:"request_id,omitempty"`
	Body       *string   `json:"body,omitempty"`
	Error      *string   `json:"error,omitempty"`
}

// RequestLogsIngestHandler recebe um lote de logs do gameServer.
func (h *Handler) RequestLogsIngestHandler(w http.ResponseWriter, r *http.Request) {
	ingestRequestLogs(w, r, h.RequestLogService)
}

func ingestRequestLogs(w http.ResponseWriter, r *http.Request, inserter requestLogInserter) {
	r.Body = http.MaxBytesReader(w, r.Body, maxIngestBody)

	var entries []ingestRequestLog
	if err := json.NewDecoder(r.Body).Decode(&entries); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			utils.SendError(w, "corpo muito grande", http.StatusRequestEntityTooLarge)
			return
		}
		utils.SendError(w, "payload inválido", http.StatusBadRequest)
		return
	}

	if len(entries) > maxIngestLines {
		utils.SendError(w, "máximo de 500 linhas por lote", http.StatusRequestEntityTooLarge)
		return
	}

	now := time.Now().UTC()
	logs := make([]*models.RequestLog, 0, len(entries))
	for i := range entries {
		logs = append(logs, sanitizeIngestLog(entries[i], now))
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := inserter.InsertBatch(ctx, logs); err != nil {
		utils.SendError(w, "erro ao gravar logs", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// sanitizeIngestLog aplica os cortes, força service=gameServer, sanea o at e
// mascara de novo o corpo/erro.
func sanitizeIngestLog(e ingestRequestLog, now time.Time) *models.RequestLog {
	at := e.At
	if at.IsZero() || at.Before(now.Add(-24*time.Hour)) || at.After(now.Add(24*time.Hour)) {
		at = now
	}

	return &models.RequestLog{
		At:         at,
		Service:    "gameServer",
		Method:     e.Method,
		Path:       truncatePtr(e.Path, truncPath),
		Route:      e.Route,
		Status:     e.Status,
		DurationMs: e.DurationMs,
		IP:         e.IP,
		UserAgent:  truncatePtr(e.UserAgent, truncUserAgent),
		UserID:     e.UserID,
		Username:   truncatePtr(e.Username, truncUsername),
		RequestID:  e.RequestID,
		Body:       maskOrTruncatePtr(e.Body, truncBody),
		Error:      maskOrTruncatePtr(e.Error, truncError),
	}
}

func truncatePtr(s *string, max int) *string {
	if s == nil {
		return nil
	}
	v := truncateUTF8Safe(*s, max)
	return &v
}

func maskOrTruncatePtr(s *string, max int) *string {
	if s == nil || *s == "" {
		return nil
	}
	v := maskOrTruncate(*s, max)
	return &v
}

// maskOrTruncate re-mascara um valor como JSON (segunda defesa) e corta em max.
// Valor não-JSON (ex.: "<N bytes>") é só cortado, sem re-serializar.
func maskOrTruncate(s string, max int) string {
	if json.Valid([]byte(s)) {
		return truncateUTF8Safe(utils.MaskBody("application/json", s), max)
	}
	return truncateUTF8Safe(s, max)
}

func truncateUTF8Safe(s string, max int) string {
	if len(s) <= max {
		return s
	}
	s = s[:max]
	for len(s) > 0 && !utf8.ValidString(s) {
		s = s[:len(s)-1]
	}
	return s
}
