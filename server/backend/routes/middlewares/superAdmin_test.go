package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
)

type fakeSuperAdminChecker struct {
	super bool
	err   error
}

func (f fakeSuperAdminChecker) IsSuperAdmin(ctx context.Context, id string) (bool, error) {
	return f.super, f.err
}

func requireReq(user structures.UserClaims) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/admin/requestLogs", nil)
	if user.Id != "" {
		ctx := context.WithValue(r.Context(), handlers.UserContextKey, user)
		r = r.WithContext(ctx)
	}
	return r
}

func TestRequireSuperAdminSemUsuario401(t *testing.T) {
	rr := httptest.NewRecorder()
	ok := requireSuperAdmin(rr, requireReq(structures.UserClaims{}), fakeSuperAdminChecker{})
	if ok {
		t.Fatal("sem usuário deveria negar")
	}
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, veio %d", rr.Code)
	}
}

func TestRequireSuperAdminComum403(t *testing.T) {
	rr := httptest.NewRecorder()
	ok := requireSuperAdmin(rr, requireReq(structures.UserClaims{Id: "u1"}), fakeSuperAdminChecker{super: false})
	if ok {
		t.Fatal("não super-admin deveria negar")
	}
	if rr.Code != http.StatusForbidden {
		t.Fatalf("esperado 403, veio %d", rr.Code)
	}
}

func TestRequireSuperAdminErroConsulta403(t *testing.T) {
	rr := httptest.NewRecorder()
	ok := requireSuperAdmin(rr, requireReq(structures.UserClaims{Id: "u1"}), fakeSuperAdminChecker{err: errors.New("db fora")})
	if ok {
		t.Fatal("erro de consulta deveria negar")
	}
	if rr.Code != http.StatusForbidden {
		t.Fatalf("esperado 403, veio %d", rr.Code)
	}
}

func TestRequireSuperAdminLibera(t *testing.T) {
	rr := httptest.NewRecorder()
	ok := requireSuperAdmin(rr, requireReq(structures.UserClaims{Id: "u1"}), fakeSuperAdminChecker{super: true})
	if !ok {
		t.Fatal("super-admin deveria liberar")
	}
}
