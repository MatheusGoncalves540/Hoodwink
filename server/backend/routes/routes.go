package routes

import (
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/routes/middlewares"
	"github.com/go-chi/chi/v5"
)

func SetupRoutes(handler *handlers.Handler) http.Handler {
	routes := chi.NewRouter()
	routes.Use(func(next http.Handler) http.Handler {
		return middlewares.RequestMiddleware(next, handler.RequestLogService)
	})
	routes.Use(middlewares.CORSMiddleware)

	// Rotas públicas
	routes.Get("/alive", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	routes.Get("/version", handler.VersionHandler)

	// Rota universal de autenticação externa
	routes.Post("/auth/external/{provider}", handler.ExternalAuthHandler)

	// Rotas protegidas com ApiKey (contrato interno: só o gameServer chama)
	routes.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return middlewares.APIKeyMiddleware(next, *handler)
		})
		r.Get("/getUserInfoById/{id}", handler.GetUserInfoByIDHandler)
		r.Post("/internal/validateToken", handler.InternalValidateTokenHandler)
		r.Post("/internal/tickets", handler.InternalTicketsHandler)
		r.Post("/internal/requestLogs", handler.RequestLogsIngestHandler)
		r.Post("/internal/replays", handler.InternalReplaysHandler)
	})

	// Rotas protegidas com JWT
	routes.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return middlewares.JWTMiddleware(next, handler)
		})
		r.Post("/auth/refresh", handler.RefreshHandler)
		r.Get("/me/matches", handler.MeMatchesHandler)
		r.Get("/replays/{matchId}", handler.ReplayHandler)
	})

	// Relatos de bug (fora do grupo JWT obrigatório global). O Lobby aceita
	// Bearer opcional; flag e PATCH exigem login válido (grupo JWT interno).
	routes.Route("/bugReports", func(r chi.Router) {
		r.Post("/", handler.BugReportLobbyHandler)
		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return middlewares.JWTMiddleware(next, handler)
			})
			r.Post("/flag", handler.BugReportFlagHandler)
			r.Patch("/{id}", handler.BugReportPatchHandler)
		})
	})

	// Rotas de super-admin (JWT + conferência de is_super_admin a cada request)
	routes.Group(func(r chi.Router) {
		r.Use(func(next http.Handler) http.Handler {
			return middlewares.JWTMiddleware(next, handler)
		})
		r.Use(func(next http.Handler) http.Handler {
			return middlewares.SuperAdminMiddleware(next, handler)
		})
		r.Get("/admin/requestLogs", handler.AdminRequestLogsHandler)
		r.Get("/admin/bugReports", handler.AdminBugReportsListHandler)
		r.Patch("/admin/bugReports/{id}", handler.AdminBugReportsPatchHandler)
	})

	return routes
}
