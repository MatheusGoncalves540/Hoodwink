package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
)

func InitSessionStore() {
	frontendURL := os.Getenv("REACT_APP_URL")
	isSecure := !strings.HasPrefix(frontendURL, "http://localhost")

	secret := os.Getenv("SESSION_SECRET")
	store := sessions.NewCookieStore([]byte(secret))
	store.Options.HttpOnly = true
	store.Options.Secure = isSecure
	store.Options.SameSite = http.SameSiteLaxMode
	gothic.Store = store
}
