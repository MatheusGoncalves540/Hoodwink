package services

import (
	"context"
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

func (s *UserService) FindOrCreateOAuthUser(ctx context.Context, email, provider, username string) (*models.User, error) {
	var user models.User
	result := s.db.WithContext(ctx).Where("email = ? AND provider = ?", email, provider).First(&user)

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
			if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, result.Error
		}
	}
	return &user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, playerID string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", playerID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// IsSuperAdmin confere a coluna users.is_super_admin a cada request (não vai no
// JWT). Erro de consulta devolve false para nunca liberar por engano.
func (s *UserService) IsSuperAdmin(ctx context.Context, playerID string) (bool, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", playerID).Error; err != nil {
		return false, err
	}
	return user.IsSuperAdmin, nil
}
