package handlers

import (
	"github.com/MatheusGoncalves540/Hoodwink/services"
)

type Handler struct {
	UserService *services.UserService
	DBService   *services.DBService
	JWTService  *services.JWTService
	// RoomService    *services.RoomService
}

func NewHandler(s *services.Services) *Handler {
	return &Handler{
		s.UserService,
		s.DBService,
		s.JWTService,
		// s.RoomService,
	}
}
