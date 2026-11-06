package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/google/uuid"
	"google.golang.org/api/idtoken"
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
		utils.SendError(w, "Provider não especificado", http.StatusBadRequest)
		return
	}

	var body ExternalAuthPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.SendError(w, "payload inválido", http.StatusBadRequest)
		return
	}

	// Validação conforme o provedor
	var email string
	var err error
	switch provider {
	case "google":
		if body.IdToken == "" {
			utils.SendError(w, "id_token obrigatório", http.StatusBadRequest)
			return
		}
		email, _, err = ValidateGoogleIDToken(body.IdToken)
	case "local":
		// Login de teste para jogar em rede local (Radmin): o Google recusa origens
		// http://<ip>, então os amigos entram só com um nome. Nunca fora de local/development.
		if env := os.Getenv("ENVIRONMENT"); env != "local" && env != "development" {
			utils.SendError(w, "Provedor não suportado", http.StatusBadRequest)
			return
		}
		body.Username = strings.TrimSpace(body.Username)
		if body.Username == "" {
			utils.SendError(w, "username obrigatório", http.StatusBadRequest)
			return
		}
		// Identidade única por login: dois amigos podem escolher o mesmo nome sem
		// virarem o mesmo jogador (o id do usuário vem do email).
		email = fmt.Sprintf("%s+%s@local.hoodwink", strings.ToLower(body.Username), uuid.NewString())
	// Futuro: case "discord": ...
	default:
		utils.SendError(w, "Provedor não suportado", http.StatusBadRequest)
		return
	}
	if err != nil {
		utils.SendError(w, "id_token inválido: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Busca/criação do usuário
	user, err := h.UserService.FindOrCreateOAuthUser(email, provider, body.Username)
	if err != nil {
		if errors.Is(err, services.ErrMissingUsername) {
			// Precisa de dados adicionais
			tempToken, _ := h.JWTService.GenerateJWT(structures.UserClaims{
				Email:    email,
				Provider: provider,
				Temp:     true,
			})
			utils.SendSuccess(w, map[string]string{"token": tempToken}, "need_additional_data")
			return
		}
		utils.SendError(w, "Erro ao salvar usuário", http.StatusInternalServerError)
		return
	}

	// Usuário existente ou criado com sucesso
	finalToken, err := h.JWTService.GenerateJWT(structures.UserClaims{
		Id:       user.ID,
		Username: user.Username,
		Provider: provider,
		Email:    user.Email,
	})
	if err != nil {
		utils.SendError(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}
	utils.SendSuccess(w, map[string]string{"token": finalToken}, "logged_in")
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
