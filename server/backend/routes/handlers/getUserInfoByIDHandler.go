package handlers

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

// Handler para pegar informações do usuário via ID externamente
// POST /getUserInfoById/{id}
type GetUserInfoByIDPayload struct {
	IdToken  string `json:"idToken"`
	Username string `json:"username,omitempty"`
}

func (h *Handler) GetUserInfoByIDHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("id")
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
