package auth

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
)

func InitSessionStore() {
	serverURL := os.Getenv("SERVER_URL")
	isSecure := !strings.HasPrefix(serverURL, "http://localhost")

	secret := os.Getenv("SESSION_SECRET")
	store := sessions.NewCookieStore([]byte(secret))
	store.Options.HttpOnly = true
	store.Options.Secure = isSecure
	store.Options.SameSite = http.SameSiteLaxMode
	gothic.Store = store
}
