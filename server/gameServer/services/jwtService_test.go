package services

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	return pub, priv
}

func signTicket(t *testing.T, priv ed25519.PrivateKey, method jwt.SigningMethod, key any, aud string, ttl time.Duration) string {
	t.Helper()
	claims := jwt.MapClaims{
		"playerId": "p1",
		"username": "ana",
		"roomId":   "r1",
		"aud":      aud,
		"exp":      time.Now().UTC().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func TestParseTokenValidEdDSA(t *testing.T) {
	pub, priv := newTestKeyPair(t)
	svc := &JWTService{publicKey: pub}

	ticket := signTicket(t, priv, jwt.SigningMethodEdDSA, priv, TicketAudience, time.Hour)
	claims, err := svc.ParseToken(ticket)
	if err != nil {
		t.Fatalf("ticket EdDSA válido deve passar: %v", err)
	}
	if claims["playerId"] != "p1" || claims["roomId"] != "r1" {
		t.Fatalf("claims inesperadas: %+v", claims)
	}
}

func TestParseTokenRejectsHS256Forged(t *testing.T) {
	pub, _ := newTestKeyPair(t)
	svc := &JWTService{publicKey: pub}

	ticket := signTicket(t, nil, jwt.SigningMethodHS256, []byte("segredo"), TicketAudience, time.Hour)
	if _, err := svc.ParseToken(ticket); err == nil {
		t.Fatal("ticket HS256 forjado deve ser recusado")
	}
}

func TestParseTokenRejectsAlgNone(t *testing.T) {
	pub, _ := newTestKeyPair(t)
	svc := &JWTService{publicKey: pub}

	ticket := signTicket(t, nil, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, TicketAudience, time.Hour)
	if _, err := svc.ParseToken(ticket); err == nil {
		t.Fatal("alg none deve ser recusado")
	}
}

func TestParseTokenRejectsLoginHS256(t *testing.T) {
	pub, _ := newTestKeyPair(t)
	svc := &JWTService{publicKey: pub}

	login := signTicket(t, nil, jwt.SigningMethodHS256, []byte("segredo-login"), "hoodwink-login", time.Hour)
	if _, err := svc.ParseToken(login); err == nil {
		t.Fatal("token de login HS256 não pode valer como ticket")
	}
}

func TestParseTicketPublicKeyInvalid(t *testing.T) {
	if _, err := parseTicketPublicKey(""); err == nil {
		t.Fatal("chave ausente deveria falhar")
	}
	if _, err := parseTicketPublicKey("!!! não é base64 !!!"); err == nil {
		t.Fatal("base64 inválido deveria falhar")
	}
}
