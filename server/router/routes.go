package router

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/router/routesFuncs"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes() http.Handler {
	r := chi.NewRouter()

	r.Get("/alive", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	r.Post("/creating-room", routesFuncs.CreateRoom)

	return r
}
