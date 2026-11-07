package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

type internalValidateTokenPayload struct {
	Token string `json:"token"`
}

type internalTicketsPayload struct {
	Token  string `json:"token"`
	RoomID string `json:"roomId"`
	IP     string `json:"ip"`
	UA     string `json:"ua"`
}

// InternalValidateTokenHandler revalida um token de login com o JWTService atual
// (HS256) e devolve id/username/provider, sem consultar o banco. Token temporário é 401.
// O provider permite o gameServer recusar criação de sala para visitante.
func (h *Handler) InternalValidateTokenHandler(w http.ResponseWriter, r *http.Request) {
	var body internalValidateTokenPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" {
		utils.SendError(w, "token não fornecido", http.StatusUnauthorized)
		return
	}

	user, err := h.JWTService.ValidateJWT(body.Token)
	if err != nil || user.Temp || user.Id == "" {
		utils.SendError(w, "token inválido", http.StatusUnauthorized)
		return
	}

	utils.SendSuccess(w, map[string]string{"id": user.Id, "username": user.Username, "provider": user.Provider}, "")
}

// InternalTicketsHandler revalida o token e assina um ticket de sala Ed25519.
// roomId vazio → 400; token inválido/vencido/temp → 401.
func (h *Handler) InternalTicketsHandler(w http.ResponseWriter, r *http.Request) {
	var body internalTicketsPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.SendError(w, "payload inválido", http.StatusBadRequest)
		return
	}
	if body.RoomID == "" {
		utils.SendError(w, "roomId vazio", http.StatusBadRequest)
		return
	}

	user, err := h.JWTService.ValidateJWT(body.Token)
	if err != nil || user.Temp || user.Id == "" {
		utils.SendError(w, "token inválido", http.StatusUnauthorized)
		return
	}

	// O ip/ua do ticket vêm do jogador (enviados pelo gameServer no corpo),
	// não do token de login; ausentes → vazios.
	user.IP = body.IP
	user.UA = body.UA

	ticket, err := h.TicketService.GenerateTicket(user, body.RoomID)
	if err != nil {
		utils.SendError(w, "erro ao gerar ticket", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, map[string]string{"ticket": ticket}, "")
}
