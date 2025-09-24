package handlers

import (
	"github.com/MatheusGoncalves540/Hoodwink/services"
)

type Handler struct {
	UserService  *services.UserService
	NewDBService *services.DBService
	// RoomService    *services.RoomService
}

func NewHandler(s *services.Services) *Handler {
	return &Handler{
		s.UserService,
		s.NewDBService,
		// s.RoomService,
	}
}
