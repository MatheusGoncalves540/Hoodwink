package services

// PostToBackend realiza um POST para a URL do backend definida nas envs ou recebida como parâmetro.
// Retorna a mensagem da resposta e um erro, se houver.
import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

// BackendService para comunicação com o backend
type BackendService struct {
	secret string
}

func NewBackendService() *BackendService {
	secret := os.Getenv("JWT_SECRET")
	return &BackendService{secret: secret}
}

// GetToBackend realiza um POST para a URL do backend definida nas envs, concatenando o path recebido.
// Headers fixos (ex: x-api-key) são definidos na requisição.
// Retorna a mensagem da resposta e um erro, se houver.
func (bs *BackendService) GetToBackend(payload any, path string) (string, error) {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		return "", errors.New("BACKEND_URL não definido no ambiente")
	}
	if len(path) > 0 && path[0] == '/' {
		backendURL += path
	} else {
		backendURL += "/" + path
	}

	req, err := http.NewRequest("GET", backendURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", os.Getenv("BACKEND_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
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
		return "", errors.New("Erro ao fazer POST para o backend: " + result.Message)
	}

	return result.Message, nil
}
