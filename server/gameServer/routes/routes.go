package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/rHandlers"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(handler *rHandlers.Handler) http.Handler {
	routes := chi.NewRouter()

	routes.Post("/getTicket/{roomId}", handler.CreateRoom)

	return routes
}
