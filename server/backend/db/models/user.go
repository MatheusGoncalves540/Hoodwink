package models

type User struct {
	ID       string `gorm:"primaryKey"`
	Email    string `gorm:"uniqueIndex"`
	Provider string
	Username string
}
