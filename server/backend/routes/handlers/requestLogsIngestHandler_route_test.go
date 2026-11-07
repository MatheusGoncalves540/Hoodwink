package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink/routes"
)

func TestInternalRequestLogsSemAPIKey(t *testing.T) {
	router := routes.SetupRoutes(newTestHandler(t))

	req := httptest.NewRequest(http.MethodPost, "/internal/requestLogs", strings.NewReader(`[]`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 sem API key, veio %d", rr.Code)
	}
}
