package handlers

import (
	"net"
	"sync"
	"time"
)

// RateLimiter é um limitador em memória com varredura (plano 06).
// ponytail: vale por pod; se o jogo escalar, troque por Redis.
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	maxKeys int
	now     func() time.Time
	entries map[string][]int64 // key -> timestamps (unix nano) dentro da janela
}

// NewRateLimiter cria um limitador de `limit` chamadas por `window`, com teto de
// `maxKeys` chaves. `now` é o relógio injetável (nil usa time.Now).
func NewRateLimiter(limit int, window time.Duration, maxKeys int, now func() time.Time) *RateLimiter {
	if now == nil {
		now = time.Now
	}
	return &RateLimiter{
		limit:   limit,
		window:  window,
		maxKeys: maxKeys,
		now:     now,
		entries: make(map[string][]int64),
	}
}

// Allow registra uma chamada para key e devolve true se estiver dentro do limite.
func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	cutoff := now.Add(-l.window).UnixNano()

	// Varredura: descarta chaves vencidas e corta timestamps antigos.
	for k, hits := range l.entries {
		kept := hits[:0]
		for _, t := range hits {
			if t >= cutoff {
				kept = append(kept, t)
			}
		}
		if len(kept) == 0 {
			delete(l.entries, k)
		} else {
			l.entries[k] = kept
		}
	}

	hits := l.entries[key]
	if len(hits) >= l.limit {
		l.entries[key] = hits
		return false
	}

	if _, exists := l.entries[key]; !exists && len(l.entries) >= l.maxKeys {
		l.evictOldest()
	}

	l.entries[key] = append(hits, now.UnixNano())
	return true
}

// evictOldest descarta a chave com o hit mais antigo para caber no teto.
func (l *RateLimiter) evictOldest() {
	var oldestKey string
	var oldest int64
	first := true
	for k, hits := range l.entries {
		if len(hits) == 0 {
			continue
		}
		last := hits[len(hits)-1]
		if first || last < oldest {
			oldest = last
			oldestKey = k
			first = false
		}
	}
	if oldestKey != "" {
		delete(l.entries, oldestKey)
	}
}

// Len devolve o número de chaves ativas (para testes e diagnóstico).
func (l *RateLimiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.entries)
}

// ipRateKey agrupa IPv6 por /64 (zera os últimos 64 bits) e mantém IPv4 como
// está; IP inválido volta como veio.
func ipRateKey(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	if v4 := parsed.To4(); v4 != nil {
		return v4.String()
	}
	v6 := parsed.To16()
	if v6 == nil {
		return ip
	}
	masked := make(net.IP, net.IPv6len)
	copy(masked, v6)
	for i := 8; i < net.IPv6len; i++ {
		masked[i] = 0
	}
	return masked.String()
}
