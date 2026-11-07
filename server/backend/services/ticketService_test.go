package services

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/golang-jwt/jwt/v5"
)

func marshalEd25519PrivateB64(t *testing.T, priv ed25519.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func marshalEd25519PublicB64(t *testing.T, pub ed25519.PublicKey) string {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshal pkix: %v", err)
	}
	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func newTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return pub, priv
}

func TestGenerateTicketAndVerify(t *testing.T) {
	pub, priv := newTestKeyPair(t)
	svc, err := newTicketService(marshalEd25519PrivateB64(t, priv), time.Hour)
	if err != nil {
		t.Fatalf("newTicketService: %v", err)
	}

	player := structures.UserClaims{Id: "p1", Username: "ana"}
	ticket, err := svc.GenerateTicket(player, "r1")
	if err != nil {
		t.Fatalf("GenerateTicket: %v", err)
	}

	parsed, err := jwt.Parse(ticket, func(token *jwt.Token) (any, error) {
		return pub, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}), jwt.WithAudience(TicketAudience))
	if err != nil || !parsed.Valid {
		t.Fatalf("ticket deve validar com a chave pública: %v", err)
	}
}

func TestGenerateTicketAnotherKeyRejects(t *testing.T) {
	_, priv := newTestKeyPair(t)
	svc, err := newTicketService(marshalEd25519PrivateB64(t, priv), time.Hour)
	if err != nil {
		t.Fatalf("newTicketService: %v", err)
	}

	ticket, err := svc.GenerateTicket(structures.UserClaims{Id: "p1", Username: "ana"}, "r1")
	if err != nil {
		t.Fatalf("GenerateTicket: %v", err)
	}

	otherPub, _ := newTestKeyPair(t)
	if _, err := jwt.Parse(ticket, func(token *jwt.Token) (any, error) {
		return otherPub, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}), jwt.WithAudience(TicketAudience)); err == nil {
		t.Fatal("ticket deve ser recusado com outra chave pública")
	}
}

func TestGenerateTicketExpiration(t *testing.T) {
	pub, priv := newTestKeyPair(t)
	svc, err := newTicketService(marshalEd25519PrivateB64(t, priv), -time.Hour)
	if err != nil {
		t.Fatalf("newTicketService: %v", err)
	}

	ticket, err := svc.GenerateTicket(structures.UserClaims{Id: "p1", Username: "ana"}, "r1")
	if err != nil {
		t.Fatalf("GenerateTicket: %v", err)
	}

	// exp já no passado: jwt.Parse deve acusar token vencido.
	if _, err := jwt.Parse(ticket, func(token *jwt.Token) (any, error) {
		return pub, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}), jwt.WithAudience(TicketAudience)); err == nil {
		t.Fatal("ticket vencido deveria ser recusado")
	}
}

func TestTicketIsNotValidLoginJWT(t *testing.T) {
	_, priv := newTestKeyPair(t)
	svc, err := newTicketService(marshalEd25519PrivateB64(t, priv), time.Hour)
	if err != nil {
		t.Fatalf("newTicketService: %v", err)
	}
	ticket, err := svc.GenerateTicket(structures.UserClaims{Id: "p1", Username: "ana"}, "r1")
	if err != nil {
		t.Fatalf("GenerateTicket: %v", err)
	}

	jwtSvc := &JWTService{secret: "segredo-de-login", expiration: time.Hour}
	if _, err := jwtSvc.ValidateJWT(ticket); err == nil {
		t.Fatal("ticket EdDSA não pode ser aceito como token de login")
	}
}

func TestNewTicketServiceInvalidBase64(t *testing.T) {
	if _, err := newTicketService("isso não é base64!!!", time.Hour); err == nil {
		t.Fatal("base64 inválido deveria falhar")
	}
}

func TestNewTicketServiceMissingKey(t *testing.T) {
	if _, err := newTicketService("", time.Hour); err == nil {
		t.Fatal("chave ausente deveria falhar")
	}
}

func TestGenerateTicketCarriesIPAndUA(t *testing.T) {
	pub, priv := newTestKeyPair(t)
	svc, err := newTicketService(marshalEd25519PrivateB64(t, priv), time.Hour)
	if err != nil {
		t.Fatalf("newTicketService: %v", err)
	}

	player := structures.UserClaims{Id: "p1", Username: "ana", IP: "26.0.0.5", UA: "hoodwink/1.0"}
	ticket, err := svc.GenerateTicket(player, "r1")
	if err != nil {
		t.Fatalf("GenerateTicket: %v", err)
	}

	parsed, err := jwt.Parse(ticket, func(token *jwt.Token) (any, error) {
		return pub, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}), jwt.WithAudience(TicketAudience))
	if err != nil || !parsed.Valid {
		t.Fatalf("ticket deve validar: %v", err)
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["ip"] != "26.0.0.5" {
		t.Fatalf("ip = %v, esperado 26.0.0.5", claims["ip"])
	}
	if claims["ua"] != "hoodwink/1.0" {
		t.Fatalf("ua = %v, esperado hoodwink/1.0", claims["ua"])
	}
	if _, ok := claims["iat"].(float64); !ok {
		t.Fatalf("iat ausente ou não numérico: %v", claims["iat"])
	}
}
