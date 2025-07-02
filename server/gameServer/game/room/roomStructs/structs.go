package roomStructs

import "time"

type GameState string

type Player struct {
	UUID       string   `json:"uuid"`
	AliveCards []string `json:"alive_cards"`
	Coins      int      `json:"coins"`
	DeadCards  []string `json:"dead_cards"`
}

type Move struct {
	PlayerUUID string      `json:"player_uuid"`
	Action     string      `json:"action"`
	TargetUUID string      `json:"target_uuid,omitempty"`
	CardUsed   string      `json:"card_used,omitempty"`
	Payload    interface{} `json:"payload,omitempty"`
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
	State              GameState `json:"state"`
	Turn               int       `json:"turn"`
	Tax                int       `json:"tax"`
	Players            []Player  `json:"players"`
	AliveDeck          []string  `json:"alive_deck"`
	DeadDeck           []string  `json:"dead_deck"`
	CurrentMove        *Move     `json:"current_move,omitempty"`
	CurrentTurnOwner   string    `json:"current_turn_owner"`
	StartTime          time.Time `json:"start_time"`
	PlayerPending      string    `json:"player_pending,omitempty"`
	PlayersWhoWantSkip []string  `json:"players_who_want_skip"`
	GameOver           bool      `json:"game_over"`
	PendingEffects     []Effect
}
