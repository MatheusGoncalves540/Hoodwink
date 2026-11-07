package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/go-chi/chi/v5"
)

type logCapture struct {
	mu   sync.Mutex
	log  *services.RequestLogService
	got  []*models.RequestLog
	once sync.Once
}

func newLogCapture(t *testing.T) *logCapture {
	t.Helper()
	c := &logCapture{}
	c.log = services.NewRequestLogServiceForTest(func(batch []*models.RequestLog) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.got = append(c.got, batch...)
	})
	t.Cleanup(func() { c.close() })
	return c
}

// close drena o canal e devolve uma cópia das linhas gravadas. É idempotente.
func (c *logCapture) close() []*models.RequestLog {
	c.once.Do(c.log.Close)
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*models.RequestLog(nil), c.got...)
}

func TestRequestMiddlewareGravaRotaAutenticada(t *testing.T) {
	c := newLogCapture(t)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.log) })
	// Simula o que o JWTMiddleware faz: preenche o *ReqInfo do contexto.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if info, ok := req.Context().Value(reqInfoContextKey).(*ReqInfo); ok {
				info.UserID = "u1"
				info.Username = "ana"
			}
			next.ServeHTTP(w, req)
		})
	})
	r.Post("/getTicket/{RoomId}", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/getTicket/r42", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("User-Agent", "hoodwink/1.0")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log, veio %d", len(got))
	}
	entry := got[0]
	if entry.UserID == nil || *entry.UserID != "u1" {
		t.Fatalf("user_id = %v, esperado u1", entry.UserID)
	}
	if entry.Username == nil || *entry.Username != "ana" {
		t.Fatalf("username = %v, esperado ana", entry.Username)
	}
	if entry.Route == nil || *entry.Route != "/getTicket/{RoomId}" {
		t.Fatalf("route = %v, esperado /getTicket/{RoomId}", entry.Route)
	}
	if entry.Path == nil || *entry.Path != "/getTicket/r42" {
		t.Fatalf("path = %v, esperado /getTicket/r42", entry.Path)
	}
}

func TestRequestMiddlewareNaoGravaAliveEInternal(t *testing.T) {
	c := newLogCapture(t)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.log) })
	r.Get("/alive", func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Post("/internal/validateToken", func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Post("/auth/external/{provider}", func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) })

	for _, path := range []string{"/alive", "/internal/validateToken"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if strings.HasPrefix(path, "/internal") {
			req = httptest.NewRequest(http.MethodPost, path, nil)
		}
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
	}

	// Apenas /auth/external deveria ter sido gravada (1 linha).
	req := httptest.NewRequest(http.MethodPost, "/auth/external/google", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log (só /auth/external), veio %d", len(got))
	}
}

func TestRequestMiddlewareHandlerLeCorpoInteiro(t *testing.T) {
	c := newLogCapture(t)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.log) })
	r.Post("/auth/external/local", func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		w.Write(body)
	})

	raw := `{"idToken":"abc","username":"ana"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/external/local", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Body.String() != raw {
		t.Fatalf("handler deveria ler o corpo inteiro %q, leu %q", raw, rr.Body.String())
	}

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log, veio %d", len(got))
	}
	body := got[0].Body
	if body == nil {
		t.Fatal("body não gravado")
	}
	if strings.Contains(*body, "abc") {
		t.Fatalf("idToken não mascarado: %q", *body)
	}
	if !strings.Contains(*body, `"idToken":"***"`) {
		t.Fatalf("idToken deveria virar ***: %q", *body)
	}
}

func TestRequestMiddlewareBugReportsSoTamanho(t *testing.T) {
	c := newLogCapture(t)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.log) })
	r.Post("/bugReports", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(io.Discard, req.Body)
		w.WriteHeader(http.StatusCreated)
	})

	raw := `{"token":"segredo","descricao":"bug"}`
	req := httptest.NewRequest(http.MethodPost, "/bugReports", bytes.NewReader([]byte(raw)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log, veio %d", len(got))
	}
	body := got[0].Body
	if body == nil {
		t.Fatal("body não gravado para /bugReports")
	}
	if *body != "<"+strconv.Itoa(len(raw))+" bytes>" {
		t.Fatalf("bugReports deveria gravar só o tamanho, veio %q", *body)
	}
}

func TestRequestMiddlewareNaoCapturaCorpoDeGet(t *testing.T) {
	c := newLogCapture(t)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.log) })
	r.Get("/game", func(w http.ResponseWriter, req *http.Request) {
		body, _ := io.ReadAll(req.Body)
		w.Write(body)
	})

	raw := `{"token":"segredo"}`
	req := httptest.NewRequest(http.MethodGet, "/game", strings.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// O handler continua lendo o corpo inteiro; só o coletor não paga a leitura.
	if rr.Body.String() != raw {
		t.Fatalf("handler deveria ler o corpo %q, leu %q", raw, rr.Body.String())
	}

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log, veio %d", len(got))
	}
	if got[0].Body != nil {
		t.Fatalf("GET não deveria capturar corpo, veio %q", *got[0].Body)
	}
}

func TestRequestMiddlewareNaoCapturaCorpoVazio(t *testing.T) {
	c := newLogCapture(t)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.log) })
	r.Post("/bugReports", func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusCreated) })

	req := httptest.NewRequest(http.MethodPost, "/bugReports", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log, veio %d", len(got))
	}
	if got[0].Body != nil {
		t.Fatalf("POST sem corpo não deveria gravar body, veio %q", *got[0].Body)
	}
}
