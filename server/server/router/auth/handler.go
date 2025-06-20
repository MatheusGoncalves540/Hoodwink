package auth

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func init() {
	key := []byte("segredo-muito-seguro")
	store := sessions.NewCookieStore(key)
	gothic.Store = store

	goth.UseProviders(
		google.New(
			"GOOGLE_CLIENT_ID",
			"GOOGLE_CLIENT_SECRET",
			"http://localhost:3000/auth/google/callback",
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
	token, err := GenerateJWT(UserClaims{Email: user.Email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Retorne o token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
