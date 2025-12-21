package roomStructs

import "time"

type Player struct {
	Id    string   `json:"id"`
	Name  string   `json:"name"`
	Cards []string `json:"cards"`
	Coins int      `json:"coins"`
	Alive bool     `json:"alive"`
}

type Room struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Password      string            `json:"password" validate:"max=24"`
	MaxPlayers    int               `json:"maxPlayers"`
	CustomMatch   bool              `json:"customMatch"`
	Turn          int               `json:"turn"`
	Tax           int               `json:"tax"`
	Players       map[string]Player `json:"players"`
	DeadDeck      []string          `json:"deadDeck"`
	CurrentPlayer string            `json:"currentPlayer"`
	State         RoomState         `json:"state"`
	PendingAction *PlayerPlay       `json:"pendingAction"`
	PendingPlayer string            `json:"pendingPlayer"`
	GameOver      bool              `json:"gameOver"`
	StartTime     time.Time         `json:"startTime"`
	Created       time.Time         `json:"created"`
}
