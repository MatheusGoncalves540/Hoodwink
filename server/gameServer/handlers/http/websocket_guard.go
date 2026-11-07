package http

import (
	"context"
	"sync"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/connmanager"
	"github.com/gorilla/websocket"
)

const (
	// wsReadLimit é o teto por mensagem WebSocket (8 KB).
	wsReadLimit = 8 << 10
	// wsPongWait é o tempo sem pong/mensagem antes de considerar a conexão morta.
	wsPongWait = 60 * time.Second
	// wsPingPeriod envia pings com folga antes do pong expirar.
	wsPingPeriod = (wsPongWait * 9) / 10
	// wsWriteWait limita cada escrita (ping incluso).
	wsWriteWait = 10 * time.Second

	// connRateLimit/connRateWindow limitam mensagens por conexão (token bucket).
	connRateLimit  = 20
	connRateWindow = 5 * time.Second
)

// tokenBucketLimiter é um token bucket em memória, por conexão, sem dependências.
// Vale por instância; o jogo tem poucas conexões simultâneas por pod.
type tokenBucketLimiter struct {
	mu       sync.Mutex
	rate     float64 // tokens por segundo
	capacity float64
	buckets  map[string]*tokenBucket
}

type tokenBucket struct {
	tokens float64
	last   time.Time
}

func newTokenBucketLimiter(rate float64, capacity int) *tokenBucketLimiter {
	return &tokenBucketLimiter{
		rate:     rate,
		capacity: float64(capacity),
		buckets:  make(map[string]*tokenBucket),
	}
}

// Allow registra uma mensagem para key e devolve true se estiver dentro do limite.
func (l *tokenBucketLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b := l.buckets[key]
	if b == nil {
		b = &tokenBucket{tokens: l.capacity, last: now}
		l.buckets[key] = b
	}

	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.last = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Delete remove o balde da conexão quando ela é encerrada, evitando vazamento.
func (l *tokenBucketLimiter) Delete(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// connRateLimiter é o limitador global de mensagens por conexão (20 / 5s).
var connRateLimiter = newTokenBucketLimiter(float64(connRateLimit)/connRateWindow.Seconds(), connRateLimit)

// configureConn aplica os limites de leitura e o refresh de deadline via pong.
func configureConn(conn *websocket.Conn) {
	conn.SetReadLimit(wsReadLimit)
	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsPongWait))
	})
}

// renewReadDeadline empurra o deadline de leitura a cada mensagem recebida.
func renewReadDeadline(conn *websocket.Conn) error {
	return conn.SetReadDeadline(time.Now().Add(wsPongWait))
}

// startPingLoop mantém a conexão viva enviando pings periódicos pelo mutex do
// ConnManager (nunca escreve direto na conexão). Encerra ao cancelar o ctx.
func startPingLoop(ctx context.Context, roomId, playerId string) {
	go func() {
		ticker := time.NewTicker(wsPingPeriod)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := connmanager.ConnManager.Ping(roomId, playerId); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
