package rHandlers

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
)

type Handler struct {
	RoomService *services.RoomService
}

func NewHandler(s *services.Services) *Handler {
	return &Handler{
		s.RoomService,
	}
}
