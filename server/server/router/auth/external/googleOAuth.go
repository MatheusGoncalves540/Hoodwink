package external

import (
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
)

func SetupGoogleOAuth() {
	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			"http://localhost:8080/auth/google/callback",
			"email", "profile",
		),
	)
}
