package services

import (
	"errors"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db}
}

func (s *UserService) FindOrCreateOAuthUser(email, provider, username string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ? AND provider = ?", email, provider).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			user = models.User{
				ID:       uuid.New().String(),
				Email:    email,
				Provider: provider,
				Username: username,
			}
			if err := s.db.Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, result.Error
		}
	}
	return &user, nil
}
