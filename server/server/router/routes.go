package router

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router/auth"
	"github.com/MatheusGoncalves540/Hoodwink/router/routesFuncs"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes() http.Handler {
	r := chi.NewRouter()

	// Rotas públicas
	r.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Rotas protegidas com JWT
	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)

		r.Post("/creating-room", routesFuncs.CreateRoom)
	})

	return r
}
