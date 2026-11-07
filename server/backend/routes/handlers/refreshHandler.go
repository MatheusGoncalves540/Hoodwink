package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
)

const (
	refreshRateLimit   = 30
	refreshRateWindow  = 10 * time.Minute
	refreshRateMaxKeys = 10000
)

// userFinder é o que o refresh precisa do banco: recarregar o usuário por id.
type userFinder interface {
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

// RefreshHandler renova o token de login: recarrega o usuário pelo id do token
// (401 se não existir), recusa token temp e devolve {token} com exp/iat/ip/ua novos.
// A validação de assinatura/exp/aud já foi feita pelo JWTMiddleware.
func (h *Handler) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	// Limite defensivo de corpo (a rota não lê body, mas reforça o teto de 16 KB).
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)

	if !h.allowRefreshRate(w, r) {
		return
	}

	writeRefresh(w, r, h.UserService, h.JWTService)
}

// allowRefreshRate limita por usuário (ou IP, sem usuário no contexto) a 30/10min.
func (h *Handler) allowRefreshRate(w http.ResponseWriter, r *http.Request) bool {
	limiter := h.RefreshIPLimiter
	key := ipRateKey(utils.ClientIP(r))
	if user, ok := r.Context().Value(UserContextKey).(structures.UserClaims); ok && user.Id != "" {
		limiter = h.RefreshUserLimiter
		key = "u:" + user.Id
	}
	if limiter == nil || limiter.Allow(key) {
		return true
	}
	utils.SendError(w, "Muitas tentativas. Tente novamente em alguns minutos.", http.StatusTooManyRequests)
	return false
}

func writeRefresh(w http.ResponseWriter, r *http.Request, finder userFinder, jwtSvc *services.JWTService) {
	user, ok := r.Context().Value(UserContextKey).(structures.UserClaims)
	if !ok || user.Temp || user.Id == "" {
		utils.SendError(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	dbUser, err := finder.GetUserByID(r.Context(), user.Id)
	if err != nil {
		utils.SendError(w, "Usuário não encontrado", http.StatusUnauthorized)
		return
	}

	token, err := jwtSvc.GenerateJWT(structures.UserClaims{
		Id:       dbUser.ID,
		Username: dbUser.Username,
		Provider: dbUser.Provider,
		Email:    dbUser.Email,
		IP:       utils.ClientIP(r),
		UA:       utils.UserAgent(r),
	})
	if err != nil {
		utils.SendError(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	utils.SendSuccess(w, map[string]string{"token": token}, "")
}
