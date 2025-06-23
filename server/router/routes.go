package router

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router/auth"
	"github.com/MatheusGoncalves540/Hoodwink/router/auth/external"
	"github.com/MatheusGoncalves540/Hoodwink/router/auth/jwtoken"
	"github.com/MatheusGoncalves540/Hoodwink/router/middlewares"
	"github.com/MatheusGoncalves540/Hoodwink/router/routesFuncs"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes() http.Handler {
	auth.InitSessionStore()
	external.SetupExternalAuths()

	routes := chi.NewRouter()
	routes.Use(middlewares.RequestMiddleware)

	// Rotas públicas
	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// OAuth
	routes.Route("/auth/{provider}", func(r chi.Router) {
		r.Get("/", external.BeginAuthHandler)
		r.Get("/callback", external.CallbackHandler)
	})
	routes.Post("/additionalUserData", external.AdditionalUserDataHandler)

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(jwtoken.JWTMiddleware)

		r.Post("/create-room", routesFuncs.CreateRoom)
	})

	return routes
}
