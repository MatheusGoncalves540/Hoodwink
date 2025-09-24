package models

import "time"

type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex"`
	Provider  string
	Username  string
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
