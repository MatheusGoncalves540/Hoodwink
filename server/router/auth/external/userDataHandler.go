package external

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink/db"
	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/router/auth/jwtoken"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
)

type AdditionalUserDataPayload struct {
	Username string `json:"username"`
}

func AdditionalUserDataHandler(w http.ResponseWriter, r *http.Request) {
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

	//user, err := h.UserService.FindOrCreateOAuthUser(email, provider, body.Username)
	var user models.User
	result := db.DB.Where("email = ? AND provider = ?", email, provider).First(&user)

	if result.Error != nil {
		if result.Error.Error() == "record not found" {
			user = models.User{
				ID:       uuid.New().String(),
				Email:    email,
				Provider: provider,
				Username: body.Username,
			}
			if err := db.DB.Create(&user).Error; err != nil {
				http.Error(w, "Erro ao criar usuário", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Erro ao buscar usuário", http.StatusInternalServerError)
			return
		}
	}

	token, err := jwtoken.GenerateJWT(jwtoken.UserClaims{
		Id:       user.ID,
		Username: user.Username,
	})

	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	frontendURL := os.Getenv("REACT_APP_URL")
	isSecure := !strings.HasPrefix(frontendURL, "http://localhost")

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
