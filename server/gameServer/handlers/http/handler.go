package http

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
)

type HTTPHandler struct {
	RoomService    *services.RoomService
	JWTService     *services.JWTService
	BackendService *services.BackendService
}

func NewHandler(s *services.Services) *HTTPHandler {
	return &HTTPHandler{
		s.RoomService,
		s.JWTService,
		s.BackendService,
	}
}
