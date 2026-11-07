package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
)

type fakeUserFinder struct {
	user *models.User
	err  error
}

func (f fakeUserFinder) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return f.user, f.err
}

func refreshTestJWT(t *testing.T) *services.JWTService {
	t.Helper()
	t.Setenv("JWT_SECRET", "segredo-login-refresh")
	t.Setenv("JWT_EXPIRATION", "24")
	return services.NewJWTService()
}

func refreshRequest(user structures.UserClaims) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	r.Header.Set("User-Agent", "hoodwink/1.0")
	r.RemoteAddr = "127.0.0.1:5000"
	ctx := context.WithValue(r.Context(), UserContextKey, user)
	return r.WithContext(ctx)
}

func decodeRefreshToken(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()
	var out struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return out.Data.Token
}

func TestWriteRefreshRejectsTemp(t *testing.T) {
	jwtSvc := refreshTestJWT(t)
	finder := fakeUserFinder{err: errors.New("não deve ser chamado")}

	rr := httptest.NewRecorder()
	writeRefresh(rr, refreshRequest(structures.UserClaims{Temp: true}), finder, jwtSvc)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 para temp, veio %d", rr.Code)
	}
}

func TestWriteRefreshRejectsDeletedUser(t *testing.T) {
	jwtSvc := refreshTestJWT(t)
	finder := fakeUserFinder{err: errors.New("não encontrado")}

	rr := httptest.NewRecorder()
	writeRefresh(rr, refreshRequest(structures.UserClaims{Id: "p1"}), finder, jwtSvc)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 para usuário apagado, veio %d", rr.Code)
	}
}

func TestAllowRefreshRatePorUsuario(t *testing.T) {
	h := &Handler{
		RefreshUserLimiter: NewRateLimiter(2, 10*time.Minute, 100, nil),
		RefreshIPLimiter:   NewRateLimiter(2, 10*time.Minute, 100, nil),
	}
	r := refreshRequest(structures.UserClaims{Id: "u1"})

	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		if !h.allowRefreshRate(rr, r) {
			t.Fatalf("chamada %d deveria passar", i+1)
		}
	}

	rr := httptest.NewRecorder()
	if h.allowRefreshRate(rr, r) {
		t.Fatal("a 3ª chamada deveria ser recusada")
	}
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("esperado 429, veio %d", rr.Code)
	}
}

func TestWriteRefreshReturnsFreshToken(t *testing.T) {
	jwtSvc := refreshTestJWT(t)
	finder := fakeUserFinder{user: &models.User{
		ID:       "p1",
		Username: "ana-nova",
		Provider: "google",
		Email:    "ana@example.com",
	}}

	rr := httptest.NewRecorder()
	writeRefresh(rr, refreshRequest(structures.UserClaims{Id: "p1", Username: "ana-antiga"}), finder, jwtSvc)
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}

	token := decodeRefreshToken(t, rr)
	if token == "" {
		t.Fatal("token ausente")
	}

	got, err := jwtSvc.ValidateJWT(token)
	if err != nil {
		t.Fatalf("token renovado deve validar: %v", err)
	}
	if got.Username != "ana-nova" || got.Email != "ana@example.com" {
		t.Fatalf("username/email não copiados do banco: %+v", got)
	}
	if got.IP != "127.0.0.1" || got.UA != "hoodwink/1.0" {
		t.Fatalf("ip/ua não gravados no refresh: %+v", got)
	}
	if got.Iat == 0 {
		t.Fatal("iat ausente no token renovado")
	}
}
