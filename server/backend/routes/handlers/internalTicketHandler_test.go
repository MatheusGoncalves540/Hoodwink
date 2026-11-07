package handlers_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/routes"
	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret = "segredo-login-teste"
	testAPIKey = "api-key-teste"
)

func newTestHandler(t *testing.T) *handlers.Handler {
	t.Helper()
	t.Setenv("JWT_SECRET", testSecret)
	t.Setenv("JWT_EXPIRATION", "24")
	t.Setenv("BACKEND_API_KEY", testAPIKey)
	t.Setenv("TICKET_EXPIRATION", "2")

	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8: %v", err)
	}
	privB64 := base64.StdEncoding.EncodeToString(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
	t.Setenv("TICKET_PRIVATE_KEY", privB64)

	ticketService, err := services.NewTicketService()
	if err != nil {
		t.Fatalf("NewTicketService: %v", err)
	}

	return &handlers.Handler{
		JWTService:    services.NewJWTService(),
		TicketService: ticketService,
		APIKey:        testAPIKey,
	}
}

func loginToken(t *testing.T, u structures.UserClaims, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"id":       u.Id,
		"username": u.Username,
		"provider": u.Provider,
		"email":    u.Email,
		"temp":     u.Temp,
		"aud":      services.LoginAudience,
		"exp":      exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("signed login token: %v", err)
	}
	return signed
}

func doPost(t *testing.T, h http.Handler, path string, body any, apiKey string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("x-api-key", apiKey)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func doPostBearer(t *testing.T, h http.Handler, path string, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestRefresh(t *testing.T) {
	router := routes.SetupRoutes(newTestHandler(t))

	t.Run("temp 401", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Email: "a@b.c", Provider: "p", Temp: true}, time.Now().Add(time.Hour))
		rr := doPostBearer(t, router, "/auth/refresh", token)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("esperado 401 para temp, veio %d", rr.Code)
		}
	})

	t.Run("vencido 401", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Id: "p1", Username: "ana", Provider: "p", Email: "a@b.c"}, time.Now().Add(-time.Hour))
		rr := doPostBearer(t, router, "/auth/refresh", token)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("esperado 401 para vencido, veio %d", rr.Code)
		}
	})
}

func TestInternalValidateToken(t *testing.T) {
	router := routes.SetupRoutes(newTestHandler(t))

	t.Run("sem api key", func(t *testing.T) {
		rr := doPost(t, router, "/internal/validateToken", map[string]string{"token": "x"}, "")
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("esperado 401, veio %d", rr.Code)
		}
	})

	t.Run("token temp", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Email: "a@b.c", Provider: "p", Temp: true}, time.Now().Add(time.Hour))
		rr := doPost(t, router, "/internal/validateToken", map[string]string{"token": token}, testAPIKey)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("esperado 401 para temp, veio %d", rr.Code)
		}
	})

	t.Run("token vencido", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Id: "p1", Username: "ana", Provider: "p", Email: "a@b.c"}, time.Now().Add(-time.Hour))
		rr := doPost(t, router, "/internal/validateToken", map[string]string{"token": token}, testAPIKey)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("esperado 401 para vencido, veio %d", rr.Code)
		}
	})

	t.Run("token válido", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Id: "p1", Username: "ana", Provider: "p", Email: "a@b.c"}, time.Now().Add(time.Hour))
		rr := doPost(t, router, "/internal/validateToken", map[string]string{"token": token}, testAPIKey)
		if rr.Code != http.StatusOK {
			t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
		}
		var out struct {
			Data struct {
				Id       string `json:"id"`
				Username string `json:"username"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if out.Data.Id != "p1" || out.Data.Username != "ana" {
			t.Fatalf("id/username inesperados: %+v", out.Data)
		}
	})
}

func TestInternalTickets(t *testing.T) {
	router := routes.SetupRoutes(newTestHandler(t))

	t.Run("roomId vazio", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Id: "p1", Username: "ana", Provider: "p", Email: "a@b.c"}, time.Now().Add(time.Hour))
		rr := doPost(t, router, "/internal/tickets", map[string]string{"token": token, "roomId": ""}, testAPIKey)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("esperado 400 para roomId vazio, veio %d", rr.Code)
		}
	})

	t.Run("token temp", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Email: "a@b.c", Provider: "p", Temp: true}, time.Now().Add(time.Hour))
		rr := doPost(t, router, "/internal/tickets", map[string]string{"token": token, "roomId": "r1"}, testAPIKey)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("esperado 401 para temp, veio %d", rr.Code)
		}
	})

	t.Run("válido", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Id: "p1", Username: "ana", Provider: "p", Email: "a@b.c"}, time.Now().Add(time.Hour))
		rr := doPost(t, router, "/internal/tickets", map[string]string{"token": token, "roomId": "r1"}, testAPIKey)
		if rr.Code != http.StatusOK {
			t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
		}
		var out struct {
			Data struct {
				Ticket string `json:"ticket"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if out.Data.Ticket == "" {
			t.Fatal("ticket vazio")
		}
	})

	t.Run("grava ip/ua do corpo", func(t *testing.T) {
		token := loginToken(t, structures.UserClaims{Id: "p1", Username: "ana", Provider: "p", Email: "a@b.c"}, time.Now().Add(time.Hour))
		rr := doPost(t, router, "/internal/tickets", map[string]string{
			"token": token, "roomId": "r1", "ip": "26.0.0.5", "ua": "hoodwink/1.0",
		}, testAPIKey)
		if rr.Code != http.StatusOK {
			t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
		}
		var out struct {
			Data struct {
				Ticket string `json:"ticket"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("decode: %v", err)
		}

		claims := jwt.MapClaims{}
		if _, _, err := jwt.NewParser().ParseUnverified(out.Data.Ticket, claims); err != nil {
			t.Fatalf("parse unverified: %v", err)
		}
		if claims["ip"] != "26.0.0.5" {
			t.Fatalf("ip = %v, esperado 26.0.0.5", claims["ip"])
		}
		if claims["ua"] != "hoodwink/1.0" {
			t.Fatalf("ua = %v, esperado hoodwink/1.0", claims["ua"])
		}
	})
}
