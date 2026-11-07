package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
)

type fakeInserter struct {
	entries []*models.RequestLog
	err     error
}

func (f *fakeInserter) InsertBatch(ctx context.Context, entries []*models.RequestLog) error {
	f.entries = append(f.entries, entries...)
	return f.err
}

func TestSanitizeIngestLog(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

	body := `{"ticket":"segredo","token":"x"}`
	route := "/getTicket/{RoomId}"
	ua := strings.Repeat("u", 300)
	username := strings.Repeat("n", 100)

	e := ingestRequestLog{
		At:        now.Add(-48 * time.Hour), // fora de ±24h → vira now
		Service:   "backend",                // forçado para gameServer
		Path:      strPtr(strings.Repeat("p", 600)),
		Route:     &route,
		UserAgent: &ua,
		Username:  &username,
		Body:      &body,
		Error:     strPtr(`{"message":"x","password":"segredo"}`),
	}

	got := sanitizeIngestLog(e, now)

	if got.Service != "gameServer" {
		t.Fatalf("service = %q, esperado gameServer", got.Service)
	}
	if !got.At.Equal(now) {
		t.Fatalf("at = %v, esperado %v", got.At, now)
	}
	if got.Path == nil || len(*got.Path) > truncPath {
		t.Fatalf("path não cortado em %d: %v", truncPath, got.Path)
	}
	if got.UserAgent == nil || len(*got.UserAgent) > truncUserAgent {
		t.Fatalf("user_agent não cortado em %d", truncUserAgent)
	}
	if got.Username == nil || len(*got.Username) > truncUsername {
		t.Fatalf("username não cortado em %d", truncUsername)
	}
	if got.Body == nil || strings.Contains(*got.Body, "segredo") {
		t.Fatalf("body deveria mascarar ticket: %v", got.Body)
	}
	if !strings.Contains(*got.Body, `"ticket":"***"`) {
		t.Fatalf("ticket deveria virar ***: %v", got.Body)
	}
	if got.Error == nil || strings.Contains(*got.Error, "segredo") {
		t.Fatalf("error deveria mascarar password: %v", got.Error)
	}
}

func TestIngestRequestLogs204(t *testing.T) {
	fake := &fakeInserter{}
	raw := `[{"service":"backend","body":"{\"ticket\":\"abc\"}"}]`
	req := httptest.NewRequest(http.MethodPost, "/internal/requestLogs", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ingestRequestLogs(rr, req, fake)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("esperado 204, veio %d: %s", rr.Code, rr.Body.String())
	}
	if len(fake.entries) != 1 {
		t.Fatalf("esperado 1 linha gravada, veio %d", len(fake.entries))
	}
	if fake.entries[0].Service != "gameServer" {
		t.Fatalf("service = %q, esperado gameServer", fake.entries[0].Service)
	}
	if fake.entries[0].Body == nil || strings.Contains(*fake.entries[0].Body, "abc") {
		t.Fatalf("body deveria mascarar ticket: %v", fake.entries[0].Body)
	}
}

func TestIngestRequestLogs413MaisDe500Linhas(t *testing.T) {
	fake := &fakeInserter{}
	var entries []ingestRequestLog
	for i := 0; i < 501; i++ {
		entries = append(entries, ingestRequestLog{Service: "gameServer", At: time.Now().UTC()})
	}
	body, _ := json.Marshal(entries)

	req := httptest.NewRequest(http.MethodPost, "/internal/requestLogs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ingestRequestLogs(rr, req, fake)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("esperado 413, veio %d", rr.Code)
	}
	if len(fake.entries) != 0 {
		t.Fatalf("não deveria gravar nada acima de 500 linhas, gravou %d", len(fake.entries))
	}
}

func TestIngestRequestLogsPayloadInvalido400(t *testing.T) {
	fake := &fakeInserter{}
	req := httptest.NewRequest(http.MethodPost, "/internal/requestLogs", strings.NewReader(`não é json`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ingestRequestLogs(rr, req, fake)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d", rr.Code)
	}
}

func strPtr(s string) *string { return &s }
