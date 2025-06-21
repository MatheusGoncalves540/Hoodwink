package router

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router/auth"
	"github.com/MatheusGoncalves540/Hoodwink/router/auth/jwt"
	"github.com/MatheusGoncalves540/Hoodwink/router/middlewares"
	"github.com/MatheusGoncalves540/Hoodwink/router/routesFuncs"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes() http.Handler {
	routes := chi.NewRouter()
	
	auth.SetupExternalAuths()

	routes.Use(middlewares.RequestMiddleware)

	// Rotas públicas
	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	routes.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		jwtToken, err := jwt.GenerateJWT(jwt.UserClaims{Email: "mrbuugames@gmail.com", Username: "Matheus Gonçalves", Level: 1})
		if err != nil {
			http.Error(w, "Erro ao gerar token de autenticação", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(jwtToken))
	})

	routes.Group(func(r chi.Router) {
		// Rotas de autenticação com provedores externos
		r.Get("/auth/{provider}", auth.GoogleLogin)
		r.Get("/auth/{provider}/callback", auth.GoogleCallback)
	})

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(jwt.JWTMiddleware)

		r.Post("/create-room", routesFuncs.CreateRoom)
	})

	return routes
}
