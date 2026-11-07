package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// backendTimeout limita as chamadas ao backend (localhost na VM).
const backendTimeout = 3 * time.Second

// BackendService para comunicação com o backend via API key.
type BackendService struct {
	apiKey string
	url    string
	client *http.Client
}

func NewBackendService() *BackendService {
	apiKey := os.Getenv("BACKEND_API_KEY")
	backendURL := os.Getenv("BACKEND_URL")

	if apiKey == "" || backendURL == "" {
		utils.LogError("⚠️ Variáveis do backend não definidas corretamente")
	}

	return &BackendService{
		apiKey: apiKey,
		url:    backendURL,
		client: &http.Client{Timeout: backendTimeout},
	}
}

func (bs *BackendService) endpoint(path string) string {
	return strings.TrimRight(bs.url, "/") + path
}

// postJSON envia um POST com x-api-key e decodifica a resposta (2xx) em out.
// Retorna o status HTTP; erros de rede/timeout são retornados como error.
// O ctx do request propaga cancelamento/timeout ao client compartilhado.
func (bs *BackendService) postJSON(ctx context.Context, path string, payload any, out any) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, bs.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", bs.apiKey)

	resp, err := bs.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 && out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return resp.StatusCode, err
		}
	} else {
		io.Copy(io.Discard, resp.Body)
	}

	return resp.StatusCode, nil
}

// ValidateToken chama POST /internal/validateToken e devolve o status já
// mapeado para o cliente: 200 (ok), 401 (token recusado) ou 503 (backend fora).
func (bs *BackendService) ValidateToken(ctx context.Context, token string) (id, username string, status int) {
	var out struct {
		Id       string `json:"id"`
		Username string `json:"username"`
	}
	st, err := bs.postJSON(ctx, "/internal/validateToken", map[string]string{"token": token}, &out)
	if err != nil {
		return "", "", http.StatusServiceUnavailable
	}
	if st == http.StatusUnauthorized {
		return "", "", http.StatusUnauthorized
	}
	if st != http.StatusOK {
		return "", "", http.StatusServiceUnavailable
	}
	return out.Id, out.Username, http.StatusOK
}

// RequestTicket chama POST /internal/tickets e devolve o status já mapeado
// para o cliente: 200 (ticket), 401 (token recusado) ou 503 (backend fora).
// ip/ua vêm do jogador (ClientIP/UserAgent) e são gravados no ticket pelo backend.
func (bs *BackendService) RequestTicket(ctx context.Context, token, roomId, ip, ua string) (ticket string, status int) {
	var out struct {
		Ticket string `json:"ticket"`
	}
	st, err := bs.postJSON(ctx, "/internal/tickets", map[string]string{
		"token":  token,
		"roomId": roomId,
		"ip":     ip,
		"ua":     ua,
	}, &out)
	if err != nil {
		return "", http.StatusServiceUnavailable
	}
	if st == http.StatusUnauthorized {
		return "", http.StatusUnauthorized
	}
	if st != http.StatusOK {
		return "", http.StatusServiceUnavailable
	}
	return out.Ticket, http.StatusOK
}

// GetToBackend realiza um GET para a URL do backend concatenando o path.
// Mantido para compatibilidade; usa o mesmo client com timeout.
func (bs *BackendService) GetToBackend(ctx context.Context, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bs.endpoint(path), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", bs.apiKey)

	resp, err := bs.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Erro ao fazer GET para o backend: " + result.Message)
	}

	return result.Message, nil
}
