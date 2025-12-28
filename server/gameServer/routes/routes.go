package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config/ctx"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/middlewares"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(routesContext *ctx.RoutesContext) http.Handler {
	routes := chi.NewRouter()
	routes.Use(middlewares.RequestMiddleware)
	routes.Use(middlewares.CORSMiddleware)

	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })

	routes.Route("/game", func(r chi.Router) {
		r.Get("/", routesContext.Handler.WebSocketHandler(routesContext.Rdb))
	})

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return middlewares.JWTBackendMiddleware(next, routesContext.Handler)
		})

		r.Post("/newRoom", routesContext.Handler.CreateCustomRoom)

		r.Post("/getTicket/{RoomId}", routesContext.Handler.GetTicket(routesContext.Rdb))
	})
	// Rotas protegidas com ApiKey
	routes.Group(func(r chi.Router) {
		// r.Use(func(next http.Handler) http.Handler {
		// 	return apiKey.APIKeyMiddleware(next, *handler)
		// })

		// Rotas para debug e monitoramento de instâncias
		r.Get("/instances/status", routesContext.Handler.GetInstancesStatus(routesContext.Rdb))
		r.Post("/instances/cleanup", routesContext.Handler.CleanupOrphanedPlayers(routesContext.Rdb))
	})

	return routes
}
