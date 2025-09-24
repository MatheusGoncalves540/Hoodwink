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
	userID := chi.URLParam(r, "id")
	if userID == "" {
		utils.SendError(w, "ID do usuário não especificado", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.GetUserByID(userID)
	if err != nil {
		utils.SendError(w, "Erro ao buscar informações do usuário", http.StatusInternalServerError)
		return
	}

	utils.SendJSON(w, http.StatusOK, user)
}
