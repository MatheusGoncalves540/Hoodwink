package eventQueue

import "time"

type Event struct {
	Type      string      `json:"type"`
	PlayerId  string      `json:"playerId"`
	RoomId    string      `json:"roomId"`
	TimeoutMs int         `json:"timeoutMs"`
	CreatedAt time.Time   `json:"createdAt"`
	Payload   interface{} `json:"payload"`
}
