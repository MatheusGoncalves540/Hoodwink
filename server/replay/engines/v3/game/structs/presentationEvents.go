package structs

import (
	"encoding/json"
	"fmt"
)

// ===== EVENTO DE APRESENTAÇÃO (animação / announcer) =====
type PresentationEvent struct {
	PlayerID              string
	Type                  GameEventType
	Payload               any
	ConfidencialPayload   any
	ConfidencialPlayerIds []string
	TimeoutField          string
}

type PresentationEventDTO struct {
	PlayerID              string          `json:"playerId"`
	Type                  GameEventType   `json:"type"`
	Payload               json.RawMessage `json:"payload"`
	ConfidencialPayload   json.RawMessage `json:"confidencialPayload,omitempty"`
	ConfidencialPlayerIds []string        `json:"confidencialPlayerIds,omitempty"`
	TimeoutField          string          `json:"timeoutField,omitempty"`
}

func NewPresentationEvent(playerID string, eventType GameEventType, payload any) PresentationEvent {
	return PresentationEvent{
		PlayerID:              playerID,
		Type:                  eventType,
		Payload:               payload,
		ConfidencialPayload:   nil,
		ConfidencialPlayerIds: []string{},
		TimeoutField:          "",
	}
}

func (e PresentationEvent) ToDTO() (PresentationEventDTO, error) {
	var payload json.RawMessage
	if e.Payload != nil {
		data, err := json.Marshal(e.Payload)
		if err != nil {
			return PresentationEventDTO{}, fmt.Errorf("falha ao serializar payload do evento %s: %w", e.Type, err)
		}
		payload = data
	}

	var confidencialPayload json.RawMessage
	if e.ConfidencialPayload != nil {
		data, err := json.Marshal(e.ConfidencialPayload)
		if err != nil {
			return PresentationEventDTO{}, fmt.Errorf("falha ao serializar payload confidencial do evento %s: %w", e.Type, err)
		}
		confidencialPayload = data
	}

	return PresentationEventDTO{
		PlayerID:              e.PlayerID,
		Type:                  e.Type,
		Payload:               payload,
		ConfidencialPayload:   confidencialPayload,
		ConfidencialPlayerIds: e.ConfidencialPlayerIds,
		TimeoutField:          e.TimeoutField,
	}, nil
}
