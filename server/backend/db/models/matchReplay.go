package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

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
	PlayerIDs   UUIDArray `gorm:"column:player_ids;type:uuid[]"`
	Data        []byte    `gorm:"column:data;type:bytea;not null"`
}

// UUIDArray é uma coluna uuid[] do Postgres. O driver devolve o array como texto
// (`{a,b}`), e database/sql não converte isso em []string sozinho — daí o Scanner.
type UUIDArray []string

func (a *UUIDArray) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*a = nil
		return nil
	case string:
		return a.parse(v)
	case []byte:
		return a.parse(string(v))
	default:
		return fmt.Errorf("UUIDArray: tipo não suportado %T", src)
	}
}

func (a UUIDArray) Value() (driver.Value, error) {
	return "{" + strings.Join(a, ",") + "}", nil
}

// parse aceita o formato de array do Postgres: {uuid1,uuid2} ou {}.
func (a *UUIDArray) parse(s string) error {
	s = strings.TrimSpace(s)
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return fmt.Errorf("UUIDArray inválido: %q", s)
	}
	body := s[1 : len(s)-1]
	if body == "" {
		*a = UUIDArray{}
		return nil
	}
	parts := strings.Split(body, ",")
	out := make(UUIDArray, len(parts))
	for i, p := range parts {
		out[i] = strings.Trim(strings.TrimSpace(p), `"`)
	}
	*a = out
	return nil
}

func (MatchReplay) TableName() string { return "match_replays" }
