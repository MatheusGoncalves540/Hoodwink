package services

import (
	// "github.com/MatheusGoncalves540/Hoodwink-gameServer/db/models"
	"github.com/redis/go-redis/v9"
)

type RoomService struct {
	db *redis.Client
}

func NewRoomService(redisClient *redis.Client) *RoomService {
	return &RoomService{redisClient}
}

// func (s *RoomService) FindOrCreateOAuthUser(email, provider, username string) (*models.Room, error) {
func (s *RoomService) FindOrCreateOAuthUser(email, provider, username string) error {
	// var room models.Room

	// return &room, nil
	return nil
}
