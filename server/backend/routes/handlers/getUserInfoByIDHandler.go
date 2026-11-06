package handlers

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/go-chi/chi/v5"
)

// Handler para pegar informações do usuário via ID externamente
// POST /getUserInfoById/{id}
type GetUserInfoByIDPayload struct {
	IdToken  string `json:"idToken"`
	Username string `json:"username,omitempty"`
}

func (h *Handler) GetUserInfoByIDHandler(w http.ResponseWriter, r *http.Request) {
	// Pega o parâmetro {id} da rota usando chi
	playerID := chi.URLParam(r, "id")
	if playerID == "" {
		utils.SendError(w, "ID do usuário não especificado", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.GetUserByID(playerID)
	if err != nil {
		utils.SendError(w, "Erro ao buscar informações do usuário", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, user, "")
}
