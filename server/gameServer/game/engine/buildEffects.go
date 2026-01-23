package engine

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func BuildEffect(dto structs.EffectDTO) (structs.Effect, error) {
	// Converte dados crús do Redis em uma instância segura de Effect.
	if dto.Cause == "" {
		return structs.Effect{}, fmt.Errorf("effect cause não pode ser vazio")
	}
	if dto.SourcePlayer == "" {
		return structs.Effect{}, fmt.Errorf("sourcePlayer não pode ser vazio para o efeito %s", dto.Cause)
	}

	effect := structs.Effect{
		Cause:        dto.Cause,
		SourcePlayer: dto.SourcePlayer,
	}

	switch dto.Cause {
	case structs.EffectContest:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.ContestPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectContestPenalty:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.ContestPenaltyPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectAssassin:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.AssassinPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectKamikaze:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.KamikazePayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectTrillionaire:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.TrillionairePayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectGreed:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.KillCardPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	default:
		return structs.Effect{}, fmt.Errorf("efeito desconhecido: %s", dto.Cause)
	}
}
