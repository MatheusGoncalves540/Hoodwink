package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestValidateTokenMapping(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		want       int
	}{
		{"backend 401 → 401", http.StatusUnauthorized, http.StatusUnauthorized},
		{"backend 500 → 503", http.StatusInternalServerError, http.StatusServiceUnavailable},
		{"backend 200 → 200", http.StatusOK, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				// Formato real do backend (utils.SendSuccess): campo sob `data`.
				w.Write([]byte(`{"data":{"id":"p1","username":"ana","provider":"google"}}`))
			}))
			defer srv.Close()

			svc := &BackendService{apiKey: "k", url: srv.URL, client: &http.Client{Timeout: time.Second}}
			id, username, provider, status := svc.ValidateToken(context.Background(), "tok")
			if status != tc.want {
				t.Fatalf("status esperado %d, veio %d", tc.want, status)
			}
			if tc.want == http.StatusOK && (id != "p1" || username != "ana" || provider != "google") {
				t.Fatalf("id/username/provider inesperados: %q %q %q", id, username, provider)
			}
		})
	}
}

func TestValidateTokenConnectionRefused(t *testing.T) {
	svc := &BackendService{apiKey: "k", url: "http://127.0.0.1:1", client: &http.Client{Timeout: time.Second}}
	if _, _, _, status := svc.ValidateToken(context.Background(), "tok"); status != http.StatusServiceUnavailable {
		t.Fatalf("conexão recusada deveria mapear 503, veio %d", status)
	}
}

func TestValidateTokenSemIdVira503(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"username":"ana"}}`))
	}))
	defer srv.Close()

	svc := &BackendService{apiKey: "k", url: srv.URL, client: &http.Client{Timeout: time.Second}}
	if _, _, _, status := svc.ValidateToken(context.Background(), "tok"); status != http.StatusServiceUnavailable {
		t.Fatalf("resposta sem id deveria mapear 503, veio %d", status)
	}
}

func TestRequestTicketMapping(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["ip"] != "26.0.0.5" || body["ua"] != "hoodwink/1.0" {
			t.Fatalf("ip/ua não enviados no corpo: %+v", body)
		}
		w.WriteHeader(http.StatusOK)
		// Formato real do backend (utils.SendSuccess): ticket sob `data`.
		w.Write([]byte(`{"data":{"ticket":"abc.def.ghi"}}`))
	}))
	defer srv.Close()

	svc := &BackendService{apiKey: "k", url: srv.URL, client: &http.Client{Timeout: time.Second}}
	ticket, status := svc.RequestTicket(context.Background(), "tok", "r1", "26.0.0.5", "hoodwink/1.0")
	if status != http.StatusOK || ticket != "abc.def.ghi" {
		t.Fatalf("esperado 200 com ticket, veio %d / %q", status, ticket)
	}
}

func TestRequestTicketVazioVira503(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":{"ticket":""}}`))
	}))
	defer srv.Close()

	svc := &BackendService{apiKey: "k", url: srv.URL, client: &http.Client{Timeout: time.Second}}
	if ticket, status := svc.RequestTicket(context.Background(), "tok", "r1", "", ""); status != http.StatusServiceUnavailable || ticket != "" {
		t.Fatalf("ticket vazio deveria mapear 503, veio %d / %q", status, ticket)
	}
}

func TestValidateTokenRespeitaContexto(t *testing.T) {
	var hit int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hit, 1)
	}))
	defer srv.Close()

	svc := &BackendService{apiKey: "k", url: srv.URL, client: &http.Client{Timeout: 3 * time.Second}}

	// ctx já cancelado: a chamada não deve chegar ao backend (prova o WithContext).
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, _, status := svc.ValidateToken(ctx, "tok")
	if status != http.StatusServiceUnavailable {
		t.Fatalf("contexto cancelado deveria mapear 503, veio %d", status)
	}
	if atomic.LoadInt32(&hit) != 0 {
		t.Fatalf("o request não deveria chegar ao servidor com ctx cancelado")
	}
}
