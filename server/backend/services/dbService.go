package services

import (
	"gorm.io/gorm"
)

type DBService struct {
	db *gorm.DB
}

func NewDBService(db *gorm.DB) *DBService {
	return &DBService{db}
}
