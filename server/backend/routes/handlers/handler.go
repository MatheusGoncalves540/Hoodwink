package handlers

import (
	"os"

	"github.com/MatheusGoncalves540/Hoodwink/services"
	"gorm.io/gorm"
)

type Handler struct {
	DB                *gorm.DB
	UserService       *services.UserService
	DBService         *services.DBService
	JWTService        *services.JWTService
	TicketService     *services.TicketService
	RequestLogService *services.RequestLogService
	ReplayService     *services.ReplayService
	// ReplayStore é a persistência dos replays (fake nos testes).
	ReplayStore replayStore
	// Chave de API usada nas rotas internas (comparada em tempo constante).
	APIKey string
	// BugReportStore e os limitadores do plano 06 (relatos de bug).
	BugReportStore bugReportStore
	BugUserLimiter *RateLimiter
	BugIPLimiter   *RateLimiter
	// Limitadores do refresh de token (30/10min por usuário/IP).
	RefreshUserLimiter *RateLimiter
	RefreshIPLimiter   *RateLimiter
	// RoomService    *services.RoomService
}

func NewHandler(s *services.Services) *Handler {
	return &Handler{
		DB:                 s.DB,
		UserService:        s.UserService,
		DBService:          s.DBService,
		JWTService:         s.JWTService,
		TicketService:      s.TicketService,
		RequestLogService:  s.RequestLogService,
		ReplayService:      s.ReplayService,
		ReplayStore:        gormReplayStore{db: s.DB},
		APIKey:             os.Getenv("BACKEND_API_KEY"),
		BugReportStore:     gormBugReportStore{db: s.DB},
		BugUserLimiter:     NewRateLimiter(bugUserRateLimit, bugRateWindow, bugRateMaxKeys, nil),
		BugIPLimiter:       NewRateLimiter(bugIPRateLimit, bugRateWindow, bugRateMaxKeys, nil),
		RefreshUserLimiter: NewRateLimiter(refreshRateLimit, refreshRateWindow, refreshRateMaxKeys, nil),
		RefreshIPLimiter:   NewRateLimiter(refreshRateLimit, refreshRateWindow, refreshRateMaxKeys, nil),
		// RoomService: s.RoomService,
	}
}
