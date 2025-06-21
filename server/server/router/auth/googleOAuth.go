package auth

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink/router/auth/jwt"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func SetupGoogleOAuth() {
	var sessionSecret = []byte(os.Getenv("SESSION_SECRET"))
	store := sessions.NewCookieStore(sessionSecret)
	gothic.Store = store

	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:8080/auth/google/callback",
			"email", "profile",
		),
	)
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Gere o JWT
	token, err := jwt.GenerateJWT(jwt.UserClaims{Email: user.Email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorne o token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
