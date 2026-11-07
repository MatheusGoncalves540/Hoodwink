package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

type shipperCapture struct {
	mu      sync.Mutex
	shipper *services.RequestLogShipper
	entries []*services.RequestLogEntry
}

func newShipperCapture(t *testing.T) *shipperCapture {
	t.Helper()
	c := &shipperCapture{}
	c.shipper = services.NewRequestLogShipperForTest(func(batch []*services.RequestLogEntry) error {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.entries = append(c.entries, batch...)
		return nil
	})
	t.Cleanup(c.shipper.Close)
	return c
}

func (c *shipperCapture) close() []*services.RequestLogEntry {
	c.shipper.Close()
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]*services.RequestLogEntry(nil), c.entries...)
}

func TestRequestMiddlewareGravaPathSemQuery(t *testing.T) {
	c := newShipperCapture(t)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.shipper) })
	r.Get("/game", func(w http.ResponseWriter, req *http.Request) { w.WriteHeader(http.StatusOK) })

	req := httptest.NewRequest(http.MethodGet, "/game?Ticket=abc", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log de /game, veio %d", len(got))
	}
	if got[0].Path == nil || *got[0].Path != "/game" {
		t.Fatalf("path = %v, esperado /game (sem query)", got[0].Path)
	}
	if got[0].Service != "gameServer" {
		t.Fatalf("service = %q, esperado gameServer", got[0].Service)
	}
}

func TestRequestMiddlewareWebSocketUpgradePassa(t *testing.T) {
	c := newShipperCapture(t)
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler { return RequestMiddleware(next, c.shipper) })
	r.Get("/game", func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte("ok"))
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/game"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("upgrade de WebSocket falhou (Hijack deve funcionar): %v", err)
	}
	// Lê a mensagem do servidor para garantir que o handler já processou.
	if _, _, err := conn.ReadMessage(); err != nil {
		t.Fatalf("leitura do WebSocket: %v", err)
	}
	conn.Close()

	// O defer do middleware enfileira a linha após o handler retornar no servidor.
	time.Sleep(50 * time.Millisecond)

	got := c.close()
	if len(got) != 1 {
		t.Fatalf("esperado 1 log do upgrade, veio %d", len(got))
	}
	if got[0].Path == nil || *got[0].Path != "/game" {
		t.Fatalf("path = %v, esperado /game", got[0].Path)
	}
}

type fakeFlusher struct {
	http.ResponseWriter
	flushed bool
}

func (f *fakeFlusher) Flush() { f.flushed = true }

func TestResponseWriterFlushDelega(t *testing.T) {
	f := &fakeFlusher{}
	rw := &responseWriter{ResponseWriter: f}
	rw.Flush()
	if !f.flushed {
		t.Fatal("Flush não delegado ao ResponseWriter subjacente")
	}
}
