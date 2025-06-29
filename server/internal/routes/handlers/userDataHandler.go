package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink/routes/auth/jwtoken"
	"github.com/markbates/goth/gothic"
)

type AdditionalUserDataPayload struct {
	Username string `json:"username"`
}

func (h *Handler) AdditionalUserDataHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := gothic.Store.Get(r, "temp-auth-session")

	email, ok1 := session.Values["email"].(string)
	provider, ok2 := session.Values["provider"].(string)

	if !ok1 || !ok2 {
		http.Error(w, "Sessão inválida", http.StatusBadRequest)
		return
	}

	var body AdditionalUserDataPayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" {
		http.Error(w, "Username inválido", http.StatusBadRequest)
		return
	}

	user, err := h.UserService.FindOrCreateOAuthUser(email, provider, body.Username)
	if err != nil {
		http.Error(w, "Erro ao salvar usuário", http.StatusInternalServerError)
		return
	}

	token, err := jwtoken.GenerateJWT(jwtoken.UserClaims{
		Id:       user.ID,
		Username: user.Username,
	})

	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	serverURL := os.Getenv("SERVER_URL")
	isSecure := !strings.HasPrefix(serverURL, "http://localhost")

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Secure:   isSecure,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged_in"}`))
}
