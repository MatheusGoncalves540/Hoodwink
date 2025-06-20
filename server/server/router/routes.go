package router

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router/auth"
	"github.com/MatheusGoncalves540/Hoodwink/router/middlewares"
	"github.com/MatheusGoncalves540/Hoodwink/router/routesFuncs"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes() http.Handler {
	routes := chi.NewRouter()
	routes.Use(middlewares.RequestMiddleware)

	// Rotas públicas
	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	routes.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		jwtToken, err := auth.GenerateJWT(auth.UserClaims{Email: "mrbuugames@gmail.com", Username: "Matheus Gonçalves", Level: 1})
		if err != nil {
			http.Error(w, "Erro ao gerar token de autenticação", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(jwtToken))
	})

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)

		r.Post("/create-room", routesFuncs.CreateRoom)
	})

	return routes
}
