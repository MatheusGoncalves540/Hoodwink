package services

import (
	"errors"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"gorm.io/gorm"
)

type DBService struct {
	db *gorm.DB
}

func NewDBService(db *gorm.DB) *DBService {
	return &DBService{db}
}

func (dbS *DBService) VerifyApiKey(apikey string) (*models.ApiKey, error) {
	var result models.ApiKey
	err := dbS.db.Where("ApiKey = ?", apikey).First(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("API Key não fornecida")
		}
		return nil, err
	}
	return &result, nil
}
