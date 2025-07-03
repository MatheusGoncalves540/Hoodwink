package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"google.golang.org/api/idtoken"

	"github.com/MatheusGoncalves540/Hoodwink/routes/auth/jwtoken"
	"github.com/MatheusGoncalves540/Hoodwink/services"
)

// Handler para autenticação universal via OAuth (Google, Discord, etc.)
// POST /auth/external/{provider}
type ExternalAuthPayload struct {
	IdToken  string `json:"idToken"`
	Username string `json:"username,omitempty"`
}

func (h *Handler) ExternalAuthHandler(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/auth/external/")
	if provider == "" {
		http.Error(w, "Provider não especificado", http.StatusBadRequest)
		return
	}

	var body ExternalAuthPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.IdToken == "" {
		http.Error(w, "id_token obrigatório", http.StatusBadRequest)
		return
	}

	// Validação do id_token conforme o provedor
	var email string
	var err error
	switch provider {
	case "google":
		email, _, err = ValidateGoogleIDToken(body.IdToken)
	// Futuro: case "discord": ...
	default:
		http.Error(w, "Provedor não suportado", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "id_token inválido: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Busca/criação do usuário
	user, err := h.UserService.FindOrCreateOAuthUser(email, provider, body.Username)
	if err != nil {
		if errors.Is(err, services.ErrMissingUsername) {
			// Precisa de dados adicionais
			tempToken, _ := jwtoken.GenerateJWT(jwtoken.UserClaims{
				Email:    email,
				Provider: provider,
				Temp:     true,
			})
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"need_additional_data","token":"` + tempToken + `"}`))
			return
		}
		http.Error(w, "Erro ao salvar usuário", http.StatusInternalServerError)
		return
	}

	// Usuário existente ou criado com sucesso
	finalToken, err := jwtoken.GenerateJWT(jwtoken.UserClaims{
		Id:       user.ID,
		Username: user.Username,
		Provider: provider,
		Email:    user.Email,
	})
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged_in","token":"` + finalToken + `"}`))
}

// Função para validação do id_token do Google
func ValidateGoogleIDToken(idToken string) (email, sub string, err error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return "", "", errors.New("GOOGLE_CLIENT_ID não configurado no .env")
	}

	payload, err := idtoken.Validate(context.Background(), idToken, clientID)
	if err != nil {
		return "", "", err
	}

	emailVal, ok := payload.Claims["email"].(string)
	if !ok || emailVal == "" {
		return "", "", errors.New("email não encontrado no id_token")
	}
	subVal, ok := payload.Claims["sub"].(string)
	if !ok || subVal == "" {
		return "", "", errors.New("sub não encontrado no id_token")
	}

	return emailVal, subVal, nil
}
