package roomStructs

import "time"

type GameState string

type Player struct {
	Id    string   `json:"id"`
	Cards []string `json:"cards"`
	Coins int      `json:"coins"`
}

type Move struct {
	PlayerId string      `json:"playerId"`
	Action   string      `json:"action"`
	TargetId string      `json:"targetId,omitempty"`
	CardUsed string      `json:"cardUsed,omitempty"`
	Payload  interface{} `json:"payload,omitempty"`
}

type Effect struct {
	Type       string // "kill", "gain_coin", etc.
	From       string
	To         string
	CardIndex  int
	Executable bool
	Reason     string
}

type Room struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Password           string    `json:"password" validate:"omitempty,max=24"`
	Created            time.Time `json:"created"`
	State              GameState `json:"state"`
	Turn               int       `json:"turn"`
	Tax                int       `json:"tax"`
	Players            []Player  `json:"players"`
	AliveDeck          []string  `json:"aliveDeck"`
	DeadDeck           []string  `json:"deadDeck"`
	CurrentMove        *Move     `json:"currentMove,omitempty"`
	CurrentTurnOwner   string    `json:"currentTurnOwner"`
	StartTime          time.Time `json:"startTime"`
	PlayerPending      string    `json:"playerPending,omitempty"`
	PlayersWhoWantSkip []string  `json:"playersWhoWantSkip"`
	GameOver           bool      `json:"gameOver"`
	PendingEffects     []Effect  `json:"pendingEffects"`
}
