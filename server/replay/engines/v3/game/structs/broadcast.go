package structs

import "encoding/json"

// BroadcastMessage encapsula uma mensagem de broadcast com metadados de roteamento
type BroadcastMessage struct {
	ToAll                 bool     `json:"toAll"`
	ConfidencialPlayerIds []string `json:"confidencialPlayerIds,omitempty"`
	Data                  any      `json:"data"`
}

// NewBroadcastMessage cria uma nova mensagem de broadcast para todos
func NewBroadcastMessage(data any) *BroadcastMessage {
	return &BroadcastMessage{
		ToAll:                 true,
		ConfidencialPlayerIds: []string{},
		Data:                  data,
	}
}

// NewSelectiveBroadcastMessage cria uma nova mensagem de broadcast seletivo
func NewSelectiveBroadcastMessage(data any, playerIds []string) *BroadcastMessage {
	return &BroadcastMessage{
		ToAll:                 false,
		ConfidencialPlayerIds: playerIds,
		Data:                  data,
	}
}

// ToJSON serializa a mensagem de broadcast para JSON
func (bm *BroadcastMessage) ToJSON() ([]byte, error) {
	return json.Marshal(bm)
}

// ParseBroadcastMessage desserializa uma mensagem de broadcast do JSON
func ParseBroadcastMessage(data []byte) (*BroadcastMessage, error) {
	var msg BroadcastMessage
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
