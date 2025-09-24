package models

type ApiKey struct {
	ApiKey     string `gorm:"type:uuid;primaryKey;uniqueIndex"`
	SystemName string `gorm:"uniqueIndex"`
}
