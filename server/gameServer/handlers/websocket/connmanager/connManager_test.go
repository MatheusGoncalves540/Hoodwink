package connmanager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func newTestManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]map[string]*connEntry),
		cancels:     make(map[string]context.CancelFunc),
		subscribed:  make(map[string]bool),
	}
}

// TestBroadcastConcurrentWrites roda várias escritas concorrentes para a mesma
// conexão e verifica, sob -race, que não há escrita concorrente no gorilla nem
// data race no snapshot/remoção do ConnManager.
func TestBroadcastConcurrentWrites(t *testing.T) {
	cm := newTestManager()
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

	added := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		cm.Add("r1", "p1", c)
		close(added)
		for {
			if _, _, err := c.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	<-added

	const writers = 8
	const perWriter = 50

	var wg sync.WaitGroup
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perWriter; j++ {
				cm.Broadcast("r1", []byte("msg"))
			}
		}()
	}
	wg.Wait()

	client.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := client.ReadMessage(); err != nil {
		t.Fatalf("leitura após broadcasts concorrentes: %v", err)
	}
}
