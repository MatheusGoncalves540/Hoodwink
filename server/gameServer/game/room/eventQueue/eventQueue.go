package eventQueue

type Event struct {
	Type          string      `json:"type"`
	PlayerUUID    string      `json:"player_uuid"`
	Payload       interface{} `json:"payload"`
	TimeoutMillis int64
}
