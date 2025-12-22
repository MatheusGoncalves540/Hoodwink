package services

// PostToBackend realiza um POST para a URL do backend definida nas envs ou recebida como parâmetro.
// Retorna a mensagem da resposta e um erro, se houver.
import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// BackendService para comunicação com o backend
type BackendService struct {
	apiKey string
	url    string
}

func NewBackendService() *BackendService {
	apiKey := os.Getenv("BACKEND_API_KEY")
	backendURL := os.Getenv("BACKEND_URL")

	if apiKey == "" || backendURL == "" {
		utils.LogDebug("⚠️ Variáveis do backend não definidas corretamente")
	}

	return &BackendService{
		apiKey: apiKey,
		url:    backendURL,
	}
}

// GetToBackend realiza um GET para a URL do backend definida nas envs, concatenando o path recebido.
// Headers fixos (ex: x-api-key) são definidos na requisição.
// Retorna a mensagem da resposta e um erro, se houver.
func (bs *BackendService) GetToBackend(path string) (string, error) {
	requestUrl := bs.url
	if len(path) > 0 && path[0] == '/' {
		requestUrl += path
	} else {
		requestUrl += "/" + path
	}

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", bs.apiKey)

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
		return "", errors.New("Erro ao fazer GET para o backend: " + result.Message)
	}

	return result.Message, nil
}
