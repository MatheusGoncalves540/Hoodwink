package structs

import (
	"fmt"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
)

func (p *PlayerPlayDTO) ValidatePayload() (any, error) {
	// Normaliza o tipo para facilitar a validação (ex: "COMMAND_SOMETHING" será tratado como "COMMAND")
	if strings.HasPrefix(string(p.Type), string(Command)) {
		return struct{}{}, nil
	}

	switch p.Type {
	case PlayChatMessage:
		var payload ChatMessagePayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para chat message: %w", err)
		}
		if payload.Msg == "" {
			return nil, fmt.Errorf("campo 'msg' é obrigatório e não pode ser vazio")
		}

		return payload, nil

	case PlayPass:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := pure.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para pass: %w", err)
		}
		return empty, nil

	case PlayContest:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := pure.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para contest: %w", err)
		}
		return empty, nil

	case PlayContestPenalty:
		var payload ContestPenaltyPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para contest penalty: %w", err)
		}
		if payload.TargetCardIndex != nil && *payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		return payload, nil

	case PlayTakeCoins:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := pure.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para take coins: %w", err)
		}
		return empty, nil

	case PlayBasicKill:
		var payload AssassinPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para kill card: %w", err)
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

	case PlaySwapCards:
		var payload SwapCardsPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para swap cards: %w", err)
		}
		if payload.TargetCardIndex == nil {
			return nil, fmt.Errorf("targetCardIndex é obrigatório")
		}
		if *payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		return payload, nil

	case PlayAssassinCard:
		var payload AssassinPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
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
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
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
		if err := pure.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para trillionaire: %w", err)
		}
		return empty, nil

	case PlayPoliticalCard:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := pure.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para political: %w", err)
		}
		return empty, nil

	case PlayRebelCard:
		if len(p.Payload) == 0 {
			return struct{}{}, nil
		}
		var empty struct{}
		if err := pure.DecodeStrictJSON(p.Payload, &empty); err != nil {
			return nil, fmt.Errorf("payload inválido para rebel: %w", err)
		}
		return empty, nil

	case PlayClairvoyantCard:
		var payload ClairvoyantPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
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
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para guardian: %w", err)
		}
		if payload.TargetPlayer == "" {
			return nil, fmt.Errorf("targetPlayer é obrigatório e não pode ser vazio")
		}
		if payload.TargetCardIndex < 0 {
			return nil, fmt.Errorf("targetCardIndex deve ser >= 0")
		}
		return payload, nil

	case PlayTricksterCard:
		var payload TricksterPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para trickster: %w", err)
		}
		if payload.TargetPlayer == "" {
			return nil, fmt.Errorf("targetPlayer é obrigatório e não pode ser vazio")
		}
		return payload, nil

	case PlayGravediggerCard:
		var payload GravediggerPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para gravedigger: %w", err)
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

	case PlayCroupierCard:
		var payload CroupierPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para croupier: %w", err)
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

	case PlayInvestorCard:
		if len(p.Payload) == 0 {
			return InvestorPayload{}, nil
		}
		var payload InvestorPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para investor: %w", err)
		}
		return payload, nil

	case PlaySelfishCard:
		if len(p.Payload) == 0 {
			return SelfishPayload{}, nil
		}
		var payload SelfishPayload
		if err := pure.DecodeStrictJSON(p.Payload, &payload); err != nil {
			return nil, fmt.Errorf("payload inválido para selfish: %w", err)
		}
		return payload, nil

	default:
		return nil, fmt.Errorf("tipo de jogada inválido: %s", p.Type)
	}
}

func ParsePlayerPlay(data []byte, playerId string) (*PlayerPlay, error) {
	var raw PlayerPlayDTO
	if err := pure.DecodeStrictJSON(data, &raw); err != nil {
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
