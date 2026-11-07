package services

import (
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/golang-jwt/jwt/v5"
)

const testLoginSecret = "segredo-de-login-teste"

func TestGenerateJWTIncludesNewClaims(t *testing.T) {
	svc := &JWTService{secret: testLoginSecret, expiration: time.Hour}

	token, err := svc.GenerateJWT(structures.UserClaims{
		Id:       "p1",
		Username: "ana",
		Provider: "google",
		Email:    "ana@example.com",
		IP:       "1.2.3.4",
		UA:       "hoodwink/1.0",
	})
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		return []byte(testLoginSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid {
		t.Fatalf("token gerado deve validar: %v", err)
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["ip"] != "1.2.3.4" {
		t.Fatalf("ip = %v, esperado 1.2.3.4", claims["ip"])
	}
	if claims["ua"] != "hoodwink/1.0" {
		t.Fatalf("ua = %v, esperado hoodwink/1.0", claims["ua"])
	}
	if _, ok := claims["iat"].(float64); !ok {
		t.Fatalf("iat ausente ou não numérico: %v", claims["iat"])
	}
}

func TestValidateJWTAcceptsTokenWithoutNewClaims(t *testing.T) {
	// Token HS256 montado à mão sem ip/ua/iat (token antigo): deve continuar válido.
	claims := jwt.MapClaims{
		"id":       "p1",
		"username": "ana",
		"provider": "google",
		"email":    "ana@example.com",
		"aud":      LoginAudience,
		"exp":      time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testLoginSecret))
	if err != nil {
		t.Fatalf("signed: %v", err)
	}

	svc := &JWTService{secret: testLoginSecret, expiration: time.Hour}
	got, err := svc.ValidateJWT(signed)
	if err != nil {
		t.Fatalf("token antigo sem ip/ua/iat deve validar: %v", err)
	}
	if got.Id != "p1" || got.IP != "" || got.UA != "" || got.Iat != 0 {
		t.Fatalf("claims inesperadas: %+v", got)
	}
}
