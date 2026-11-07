package handlers

import (
	"testing"
	"time"
)

func TestRateLimiterLimit(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	l := NewRateLimiter(2, 10*time.Minute, 100, func() time.Time { return now })

	if !l.Allow("a") || !l.Allow("a") {
		t.Fatal("as duas primeiras chamadas deveriam passar")
	}
	if l.Allow("a") {
		t.Fatal("a terceira chamada deveria ser recusada")
	}
}

func TestRateLimiterChavesSeparadas(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	l := NewRateLimiter(1, 10*time.Minute, 100, func() time.Time { return now })

	if !l.Allow("a") {
		t.Fatal("primeira chamada de a deveria passar")
	}
	if !l.Allow("b") {
		t.Fatal("chave b tem contador próprio e deveria passar")
	}
	if l.Allow("a") {
		t.Fatal("segunda chamada de a deveria ser recusada")
	}
}

func TestRateLimiterLiberaDepoisDaJanela(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	l := NewRateLimiter(1, 10*time.Minute, 100, func() time.Time { return now })

	if !l.Allow("a") {
		t.Fatal("primeira chamada deveria passar")
	}
	if l.Allow("a") {
		t.Fatal("segunda chamada deveria ser recusada")
	}

	now = now.Add(11 * time.Minute)
	if !l.Allow("a") {
		t.Fatal("depois da janela a chave deveria liberar")
	}
}

func TestRateLimiterVarreduraLimpaChaveVencida(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	l := NewRateLimiter(1, 10*time.Minute, 100, func() time.Time { return now })

	l.Allow("a")
	if l.Len() != 1 {
		t.Fatalf("Len() = %d, esperado 1", l.Len())
	}

	now = now.Add(11 * time.Minute)
	l.Allow("b")
	if l.Len() != 1 {
		t.Fatalf("depois da varredura, Len() = %d, esperado 1", l.Len())
	}
}

func TestRateLimiterTetoEvictaMaisAntigo(t *testing.T) {
	now := time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)
	l := NewRateLimiter(10, time.Hour, 2, func() time.Time { return now })

	l.Allow("a")
	l.Allow("b")
	if !l.Allow("c") {
		t.Fatal("nova chave deveria evictar a mais antiga e passar")
	}
	if l.Len() > 2 {
		t.Fatalf("teto ultrapassado: Len() = %d", l.Len())
	}
}

func TestIPRateKeyIPv4(t *testing.T) {
	if got := ipRateKey("192.168.0.1"); got != "192.168.0.1" {
		t.Fatalf("ipRateKey ipv4 = %q", got)
	}
}

func TestIPRateKeyIPv6AgrupaPor64(t *testing.T) {
	a := ipRateKey("2001:db8:0:0:1:2:3:4")
	b := ipRateKey("2001:db8:0:0:9:8:7:6")
	if a != b {
		t.Fatalf("mesmo /64 deveria agrupar: %q != %q", a, b)
	}
	c := ipRateKey("2001:db8:0:1::1")
	if a == c {
		t.Fatalf("/64 diferentes não deveriam agrupar: %q", a)
	}
}

func TestIPRateKeyInvalidoVoltaComoVeio(t *testing.T) {
	if got := ipRateKey("lixo"); got != "lixo" {
		t.Fatalf("ip inválido deveria voltar como veio, veio %q", got)
	}
}
