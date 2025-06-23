package router

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router/auth/external"
	"github.com/MatheusGoncalves540/Hoodwink/router/auth/jwtoken"
	"github.com/MatheusGoncalves540/Hoodwink/router/middlewares"
	"github.com/MatheusGoncalves540/Hoodwink/router/routesFuncs"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes() http.Handler {
	routes := chi.NewRouter()

	external.SetupExternalAuths()

	routes.Use(middlewares.RequestMiddleware)

	// Rotas públicas
	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	//TODO remover essa rota
	routes.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		jwtToken, err := jwtoken.GenerateJWT(jwtoken.UserClaims{Email: "mrbuugames@gmail.com", Username: "Matheus Gonçalves", Level: 1})
		if err != nil {
			http.Error(w, "Erro ao gerar token de autenticação", http.StatusInternalServerError)
			return
		}

		w.Write([]byte(jwtToken))
	})

	routes.Group(func(r chi.Router) {
		// Rotas de autenticação com provedores externos
		r.Get("/auth/{provider}", external.GoogleLogin)
		r.Get("/auth/{provider}/callback", external.GoogleCallback)
	})

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(jwtoken.JWTMiddleware)

		r.Post("/create-room", routesFuncs.CreateRoom)
	})

	return routes
}
