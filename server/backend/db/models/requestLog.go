package models

import "time"

// RequestLog é uma linha ativa do coletor de requests (plano 04).
// Campos anuláveis usam ponteiro para preservar NULL no banco.
type RequestLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement"`
	At         time.Time `gorm:"not null;default:now()"`
	Service    string    `gorm:"not null"`
	Method     *string
	Path       *string
	Route      *string
	Status     *int
	DurationMs *int    `gorm:"column:duration_ms"`
	IP         *string `gorm:"column:ip"`
	UserAgent  *string `gorm:"column:user_agent"`
	UserID     *string `gorm:"column:user_id;type:uuid"`
	Username   *string
	RequestID  *string `gorm:"column:request_id;type:uuid"`
	Body       *string
	Error      *string `gorm:"column:error"`
}

func (RequestLog) TableName() string { return "request_logs" }

// RequestLogArchive mantém o id original da linha movida pela retenção.
type RequestLogArchive struct {
	ID         int64     `gorm:"primaryKey;autoIncrement:false"`
	At         time.Time `gorm:"not null;default:now()"`
	Service    string    `gorm:"not null"`
	Method     *string
	Path       *string
	Route      *string
	Status     *int
	DurationMs *int    `gorm:"column:duration_ms"`
	IP         *string `gorm:"column:ip"`
	UserAgent  *string `gorm:"column:user_agent"`
	UserID     *string `gorm:"column:user_id;type:uuid"`
	Username   *string
	RequestID  *string `gorm:"column:request_id;type:uuid"`
	Body       *string
	Error      *string `gorm:"column:error"`
}

func (RequestLogArchive) TableName() string { return "request_logs_archive" }
