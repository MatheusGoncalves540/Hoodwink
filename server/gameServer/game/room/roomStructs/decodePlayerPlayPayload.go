package roomStructs

import (
	"encoding/json"
	"fmt"
)

type PlayerPlayPayload struct {
	Type     TypePlayerPlays `json:"type"`
	PlayerID string          `json:"playerId"`
	Payload  json.RawMessage `json:"payload"`
}

//

func (p *PlayerPlayPayload) DecodePayload() (any, error) {
	switch p.Type {
	case PlayAssassinCard:
		var payload AssassinPayload
		if err := json.Unmarshal(p.Payload, &payload); err != nil {
			return nil, err
		}
		return payload, nil

	// case PlayGuardianCard:
	// 	var payload GuardianPayload
	// 	if err := json.Unmarshal(p.Payload, &payload); err != nil {
	// 		return nil, err
	// 	}
	// 	return payload, nil

	default:
		return nil, fmt.Errorf("tipo de jogada inválido: %s", p.Type)
	}
}

//

func ParsePlayerPlay(data []byte, playerId string) (*PlayerPlay, error) {
	var raw PlayerPlayPayload
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("json inválido: %w", err)
	}

	raw.PlayerID = playerId

	payload, err := raw.DecodePayload()
	if err != nil {
		return nil, err
	}

	return &PlayerPlay{
		Type:     raw.Type,
		PlayerId: raw.PlayerID,
		Payload:  payload,
	}, nil
}
