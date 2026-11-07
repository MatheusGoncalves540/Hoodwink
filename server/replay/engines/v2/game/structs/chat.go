package structs

import "time"

type ChatMessage struct {
	PlayerName string    `json:"playerName"`
	Timestamp  time.Time `json:"timestamp"`
	Message    string    `json:"message"`
}
