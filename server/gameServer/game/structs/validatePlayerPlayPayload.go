package structs

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func (p *PlayerPlayDTO) ValidatePayload() (any, error) {
	switch p.Type {
	case PlayContest:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := utils.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para contest: %w", err)
		}
		return empty, nil

	case PlayContestPenalty:
		var payload ContestPenaltyPayload
		if err := utils.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para contest penalty: %w", err)
		}
		if payload.TargetCardIndex != nil && *payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		return payload, nil

	case PlayAssassinCard:
		var payload AssassinPayload
		if err := utils.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para assassin: %w", err)
		}
		if payload.TargetPlayer == nil || *payload.TargetPlayer == "" {
			return nil, fmt.Errorf("targetPlayer é obrigatório e não pode ser vazio")
		}
		if payload.TargetCardIndex == nil {
			return nil, fmt.Errorf("targetCardIndex é obrigatório")
		}
		if *payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		return payload, nil

	case PlayKamikazeCard:
		var payload KamikazePayload
		if err := utils.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para kamikaze: %w", err)
		}
		if payload.TargetPlayer == nil || *payload.TargetPlayer == "" {
			return nil, fmt.Errorf("targetPlayer é obrigatório e não pode ser vazio")
		}
		if payload.TargetCardIndex == nil {
			return nil, fmt.Errorf("targetCardIndex é obrigatório")
		}
		if *payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		if payload.TargetAllyCardIndex != nil && *payload.TargetAllyCardIndex < 0 {
			return nil, fmt.Errorf("targetAllyCardIndex deve ser >= 0")
		}
		return payload, nil

	case PlayTrillionaireCard:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := utils.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para trillionaire: %w", err)
		}
		return empty, nil

	case PlayPoliticalCard:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := utils.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para political: %w", err)
		}
		return empty, nil

	case PlayRebelCard:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := utils.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para rebel: %w", err)
		}
		return empty, nil

	case PlayClairvoyantCard:
		var payload ClairvoyantPayload
		if err := utils.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para clairvoyant: %w", err)
		}
		if payload.TargetPlayer == "" {
			return nil, fmt.Errorf("targetPlayer é obrigatório e não pode ser vazio")
		}
		if payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		return payload, nil

	case PlayGuardianCard:
		var payload GuardianPayload
		if err := utils.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para guardian: %w", err)
		}
		if payload.TargetPlayer == "" {
			return nil, fmt.Errorf("targetPlayer é obrigatório e não pode ser vazio")
		}
		if payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		return payload, nil

	default:
		return nil, fmt.Errorf("tipo de jogada inválido: %s", p.Type)
	}
}

func ParsePlayerPlay(data []byte, playerId string) (*PlayerPlay, error) {
	var raw PlayerPlayDTO
	if err := utils.DecodeStrictJSON(data, &raw); err != nil {
		return nil, err
	}

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
