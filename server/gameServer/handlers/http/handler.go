package http

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
)

type HTTPHandler struct {
	RoomService    *services.RoomService
	JWTService     *services.JWTService
	BackendService *services.BackendService
	RulesRegistry  *rules.Registry
}

func NewHandler(s *services.Services, r *rules.Registry) *HTTPHandler {
	return &HTTPHandler{
		s.RoomService,
		s.JWTService,
		s.BackendService,
		r,
	}
}
