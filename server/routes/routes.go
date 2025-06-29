package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/routes/auth"
	"github.com/MatheusGoncalves540/Hoodwink/routes/auth/jwtoken"
	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/routes/middlewares"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(handler *handlers.Handler) http.Handler {
	auth.InitSessionStore()
	handlers.SetupExternalAuths()

	routes := chi.NewRouter()
	routes.Use(middlewares.RequestMiddleware)

	// Rotas públicas
	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	a
	// OAuth
	routes.Route("/auth/{provider}", func(r chi.Router) {
		r.Get("/", handler.BeginAuthHandler)
		r.Get("/callback", handler.CallbackHandler)
	})
	routes.Post("/additionalUserData", handler.AdditionalUserDataHandler)

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(jwtoken.JWTMiddleware)

		r.Post("/create-room", handler.CreateRoom)
	})

	return routes
}
