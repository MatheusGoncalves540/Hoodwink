package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/websocket"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/rHandlers"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(handler *rHandlers.Handler, rdb *redis.Client) http.Handler {
	routes := chi.NewRouter()

	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })

	routes.Post("/newRoom", handler.CreateRoom)

	routes.Post("/getTicket/{RoomId}", handler.GetTicket)

	routes.Route("/game", func(r chi.Router) {
		r.Get("/", websocket.WebSocketHandler(rdb))
	})

	return routes
}
