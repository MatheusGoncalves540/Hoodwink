package engine

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// BuildEffect é responsável por construir uma instância de Effect a partir de um EffectDTO, validando o payload e garantindo que os dados estejam corretos antes de criar o efeito. Ele é usado para converter os dados brutos do Redis em uma estrutura segura e utilizável dentro do motor de jogo.
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
	case structs.EffectCardKilled:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.KillCardPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectEarnCoins:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.EarnCoinsPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectStealCoins:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.StealCoinsPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

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

	case structs.EffectChangedCard:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.ChangeCardPayload
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

	case structs.EffectPolitical:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.PoliticalPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectRebel:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.RebelPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectClairvoyant:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.ClairvoyantPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectGuardian:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.GuardianPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectTrickster:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.TricksterPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectGravedigger:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.GravediggerPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	case structs.EffectCroupier:
		if len(dto.Payload) == 0 {
			return structs.Effect{}, fmt.Errorf("payload ausente para o efeito %s", dto.Cause)
		}
		var payload structs.CroupierPayload
		if err := utils.DecodeStrictJSON(dto.Payload, &payload); err != nil {
			return structs.Effect{}, fmt.Errorf("payload inválido para o efeito %s: %w", dto.Cause, err)
		}
		effect.Payload = payload
		return effect, nil

	default:
		return structs.Effect{}, fmt.Errorf("efeito desconhecido: %s", dto.Cause)
	}
}
