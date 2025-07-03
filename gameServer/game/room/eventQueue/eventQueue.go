package eventQueue

type Event struct {
	Type          string      `json:"type"`
	PlayerId      string      `json:"playerId"`
	Payload       interface{} `json:"payload"`
	TimeoutMillis int64
}
