package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// newGuardServer sobe um servidor WebSocket mínimo que usa os mesmos guardas do
// handler real (configureConn + renovação de deadline + rate limit), ecoando cada
// mensagem aceita. Devolve o cliente já conectado.
func newGuardServer(t *testing.T, limiter *tokenBucketLimiter) (*httptest.Server, *websocket.Conn) {
	t.Helper()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		configureConn(conn)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := renewReadDeadline(conn); err != nil {
				return
			}
			if !limiter.Allow("p1") {
				conn.Close()
				return
			}
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })

	return srv, client
}

func TestWebSocketReadLimitFechaMensagemGrande(t *testing.T) {
	// Limite alto: quem derruba a conexão deve ser o SetReadLimit(8 KB), não o bucket.
	_, client := newGuardServer(t, newTokenBucketLimiter(1000, 1000))

	big := make([]byte, 9<<10)
	for i := range big {
		big[i] = 'x'
	}
	if err := client.WriteMessage(websocket.TextMessage, big); err != nil {
		t.Fatalf("write: %v", err)
	}

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := client.ReadMessage(); err == nil {
		t.Fatal("esperava fechamento por mensagem acima de 8 KB")
	}
}

func TestWebSocketRateLimitNaoDisparaEmUsoNormal(t *testing.T) {
	// Mesmo limite do handler (20 / 5s): 5 mensagens seguidas não devem fechar.
	limiter := newTokenBucketLimiter(float64(connRateLimit)/connRateWindow.Seconds(), connRateLimit)
	_, client := newGuardServer(t, limiter)

	for i := 0; i < 5; i++ {
		if err := client.WriteMessage(websocket.TextMessage, []byte("msg")); err != nil {
			t.Fatalf("write %d: %v", i, err)
		}
	}

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	for i := 0; i < 5; i++ {
		if _, _, err := client.ReadMessage(); err != nil {
			t.Fatalf("leitura %d falhou (conexão fechada indevidamente): %v", i, err)
		}
	}
}

func TestTokenBucketExcedeLimite(t *testing.T) {
	l := newTokenBucketLimiter(float64(connRateLimit)/connRateWindow.Seconds(), connRateLimit)

	for i := 0; i < connRateLimit; i++ {
		if !l.Allow("k") {
			t.Fatalf("mensagem %d deveria passar", i+1)
		}
	}
	if l.Allow("k") {
		t.Fatal("a 21ª mensagem deveria ser recusada")
	}
}

func TestTokenBucketDeleteLiberaNovaChave(t *testing.T) {
	l := newTokenBucketLimiter(float64(connRateLimit)/connRateWindow.Seconds(), connRateLimit)

	if !l.Allow("k") {
		t.Fatal("primeira mensagem deveria passar")
	}
	l.Delete("k")
	if !l.Allow("k") {
		t.Fatal("após Delete, a chave deveria voltar ao balde cheio")
	}
}
