package models

import "time"

// MatchReplay é um replay assinado de uma partida encerrada (plano 05, T9).
//
// player_ids guarda os UUIDs dos jogadores na ordem dos índices do cabeçalho .hwr
// (o formato .hwr congela só índice+nome, sem UUID; os UUIDs chegam no corpo do
// POST /internal/replays e são conferidos contra a contagem do cabeçalho).
type MatchReplay struct {
	MatchID       string `gorm:"column:match_id;type:uuid;primaryKey"`
	RoomID        string `gorm:"column:room_id;type:text"`
	EngineVersion int    `gorm:"column:engine_version;not null"`
	// GameVersion é a versão do jogo (VERSION), independente da versão do motor.
	GameVersion string    `gorm:"column:game_version;not null"`
	StartedAt   time.Time `gorm:"column:started_at;not null"`
	EndedAt     time.Time `gorm:"column:ended_at;not null"`
	WinnerID    *string   `gorm:"column:winner_id;type:uuid"`
	PlayerIDs   []string  `gorm:"column:player_ids;type:uuid[]"`
	Data        []byte    `gorm:"column:data;type:bytea;not null"`
}

func (MatchReplay) TableName() string { return "match_replays" }
