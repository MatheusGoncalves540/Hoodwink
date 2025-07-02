package services

import (
	// "github.com/MatheusGoncalves540/Hoodwink-gameServer/db/models"
	"gorm.io/gorm"
)

type RoomService struct {
	db *gorm.DB
}

func NewRoomService(db *gorm.DB) *RoomService {
	return &RoomService{db}
}

// func (s *RoomService) FindOrCreateOAuthUser(email, provider, username string) (*models.Room, error) {
func (s *RoomService) FindOrCreateOAuthUser(email, provider, username string) error {
	// var room models.Room

	// return &room, nil
	return nil
}
