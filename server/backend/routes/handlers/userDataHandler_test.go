package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func postExternalAuth(provider, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/auth/external/"+provider, strings.NewReader(body))
	rr := httptest.NewRecorder()
	(&Handler{}).ExternalAuthHandler(rr, req)
	return rr
}

func TestGuestLoginAceitoEmProducaoEValidaNome(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")

	// Nome vazio para antes de tocar no banco: prova que o provedor passou pela
	// checagem de ambiente e chegou na validação do nome.
	rr := postExternalAuth("guest", `{"username":"  "}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400 para nome vazio, veio %d: %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "username obrigatório") {
		t.Fatalf("esperado erro de nome, veio: %s", rr.Body.String())
	}
}

func TestLocalLoginContinuaSoEmDesenvolvimento(t *testing.T) {
	t.Setenv("ENVIRONMENT", "production")

	rr := postExternalAuth("local", `{"username":"ana"}`)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400 para local em produção, veio %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Provedor não suportado") {
		t.Fatalf("esperado provedor recusado, veio: %s", rr.Body.String())
	}
}
