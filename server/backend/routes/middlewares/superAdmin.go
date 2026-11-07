package middlewares

import (
	"context"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink/routes/handlers"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

// superAdminChecker é o que o middleware precisa do UserService.
type superAdminChecker interface {
	IsSuperAdmin(ctx context.Context, id string) (bool, error)
}

// SuperAdminMiddleware confere users.is_super_admin a cada request (não vai no
// JWT): sem usuário no contexto → 401; não super-admin → 403.
func SuperAdminMiddleware(next http.Handler, h *handlers.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requireSuperAdmin(w, r, h.UserService) {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireSuperAdmin(w http.ResponseWriter, r *http.Request, checker superAdminChecker) bool {
	user, ok := r.Context().Value(handlers.UserContextKey).(structures.UserClaims)
	if !ok || user.Id == "" {
		utils.SendError(w, "Token inválido", http.StatusUnauthorized)
		return false
	}

	super, err := checker.IsSuperAdmin(r.Context(), user.Id)
	if err != nil || !super {
		utils.SendError(w, "Acesso negado", http.StatusForbidden)
		return false
	}
	return true
}
