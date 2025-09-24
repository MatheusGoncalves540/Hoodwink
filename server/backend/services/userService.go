package services

import (
	"errors"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

var ErrMissingUsername = errors.New("username obrigatório para novo usuário")

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
			if username == "" {
				return nil, ErrMissingUsername
			}
			uuidv7, err := uuid.NewV7()
			if err != nil {
				return nil, err
			}
			user = models.User{
				ID:       uuidv7.String(),
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

func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
