package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/middlewares"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/routeHandlers"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(handler *routeHandlers.Handler, rdb *redis.Client) http.Handler {
	routes := chi.NewRouter()
	routes.Use(middlewares.RequestMiddleware)
	routes.Use(middlewares.CORSMiddleware)

	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })

	routes.Route("/game", func(r chi.Router) {
		r.Get("/", handler.WebSocketHandler(rdb))
	})

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return middlewares.JWTBackendMiddleware(next, handler)
		})

		r.Post("/newRoom", handler.CreateCustomRoom)

		r.Post("/getTicket/{RoomId}", handler.GetTicket(rdb))
	})
	// Rotas protegidas com ApiKey
	routes.Group(func(r chi.Router) {
		// r.Use(func(next http.Handler) http.Handler {
		// 	return apiKey.APIKeyMiddleware(next, *handler)
		// })

		// Rotas para debug e monitoramento de instâncias
		r.Get("/instances/status", handler.GetInstancesStatus(rdb))
		r.Post("/instances/cleanup", handler.CleanupOrphanedPlayers(rdb))
	})

	return routes
}
