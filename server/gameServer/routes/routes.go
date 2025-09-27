package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/middlewares"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/rHandlers"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(handler *rHandlers.Handler, rdb *redis.Client) http.Handler {
	routes := chi.NewRouter()
	routes.Use(middlewares.RequestMiddleware)
	routes.Use(middlewares.CORSMiddleware)

	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })

	routes.Post("/newRoom", handler.CreateRoom)

	routes.Post("/getTicket/{RoomId}", handler.GetTicket(rdb))

	// Rotas para debug e monitoramento de instâncias
	routes.Get("/instances/status", handler.GetInstancesStatus(rdb))
	routes.Post("/instances/cleanup", handler.CleanupOrphanedPlayers(rdb))

	routes.Route("/game", func(r chi.Router) {
		r.Get("/", handler.WebSocketHandler(rdb))
	})

	return routes
}
