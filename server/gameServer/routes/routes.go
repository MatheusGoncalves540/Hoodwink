package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/rHandlers"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(handler *rHandlers.Handler) http.Handler {
	routes := chi.NewRouter()

	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })

	routes.Post("/getTicket/{RoomId}", handler.CreateRoom)

	return routes
}
