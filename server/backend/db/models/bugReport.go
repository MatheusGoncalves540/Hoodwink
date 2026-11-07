package models

import (
	"encoding/json"
	"time"
)

// BugReport é um relato marcado na partida ou enviado pelo Lobby (plano 06).
type BugReport struct {
	ID          int64           `gorm:"primaryKey;autoIncrement"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"`
	UserID      *string         `gorm:"column:user_id;type:uuid"`
	Username    *string         `gorm:"column:username"`
	Source      string          `gorm:"not null"`
	MatchID     *string         `gorm:"column:match_id;type:uuid"`
	Turns       []int32         `gorm:"column:turns;type:int[]"`
	Category    *string         `gorm:"column:category"`
	Description *string         `gorm:"column:description"`
	Context     json.RawMessage `gorm:"column:context;type:jsonb"`
	IP          *string         `gorm:"column:ip"`
	UserAgent   *string         `gorm:"column:user_agent"`
	Status      string          `gorm:"not null;default:novo"`
}

func (BugReport) TableName() string { return "bug_reports" }
