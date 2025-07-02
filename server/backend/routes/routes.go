package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/routes/auth/jwtoken"
	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/routes/middlewares"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(handler *handlers.Handler) http.Handler {
	routes := chi.NewRouter()
	routes.Use(middlewares.RequestMiddleware)
	routes.Use(middlewares.CORSMiddleware)

	// Rotas públicas
	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Rota universal de autenticação externa
	routes.Post("/auth/external/{provider}", handler.ExternalAuthHandler)

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(jwtoken.JWTMiddleware)
	})

	return routes
}
