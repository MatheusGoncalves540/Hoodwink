package models

type ApiKey struct {
	Apikey     string `gorm:"type:uuid;primaryKey;uniqueIndex"`
	SystemName string `gorm:"uniqueIndex"`
}
