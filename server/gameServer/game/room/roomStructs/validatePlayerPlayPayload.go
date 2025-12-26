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

func (p *PlayerPlayPayload) ValidatePayload() (any, error) {
	switch p.Type {
	case PlayAssassinCard:
		var payload AssassinPayload
		if err := json.Unmarshal(p.Payload, &payload); err != nil {
			return nil, err
		}
		// Validação: campos obrigatórios devem estar presentes
		if payload.TargetPlayer == nil || *payload.TargetPlayer == "" {
			return nil, fmt.Errorf("targetPlayer é obrigatório e não pode ser vazio")
		}
		if payload.TargetCard == nil {
			return nil, fmt.Errorf("targetCard é obrigatório")
		}
		if *payload.TargetCard < 0 {
			return nil, fmt.Errorf("targetCard deve ser >= 0")
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

	// Validação: type é obrigatório
	if raw.Type == "" {
		return nil, fmt.Errorf("campo 'type' está faltando ou vazio")
	}

	raw.PlayerID = playerId

	payload, err := raw.ValidatePayload()
	if err != nil {
		return nil, err
	}

	return &PlayerPlay{
		Type:     raw.Type,
		PlayerId: raw.PlayerID,
		Payload:  payload,
	}, nil
}
