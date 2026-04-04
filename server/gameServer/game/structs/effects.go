package structs

import (
	"encoding/json"
	"fmt"
)

type EffectCause string

type EffectDTO struct {
	Cause        EffectCause     `json:"cause"`
	SourcePlayer string          `json:"sourcePlayer"`
	Payload      json.RawMessage `json:"payload"`
}

// ===== EFEITO (precisa ser executado) =====
type Effect struct {
	Cause        EffectCause
	SourcePlayer string
	Payload      any
}

func NewEffect(cause EffectCause, sourcePlayer string, payload any) Effect {
	return Effect{Cause: cause, SourcePlayer: sourcePlayer, Payload: payload}
}

func (e Effect) ToDTO() (EffectDTO, error) {
	var payload json.RawMessage
	if e.Payload != nil {
		data, err := json.Marshal(e.Payload)
		if err != nil {
			return EffectDTO{}, fmt.Errorf("falha ao serializar payload do efeito %s: %w", e.Cause, err)
		}
		payload = data
	}

	return EffectDTO{
		Cause:        e.Cause,
		SourcePlayer: e.SourcePlayer,
		Payload:      payload,
	}, nil
}

// Tipos de efeitos
const (
	EffectContest        EffectCause = "CONTEST"
	EffectContestPenalty EffectCause = "CONTEST_PENALTY"
	EffectGreed          EffectCause = "GREED"
	EffectEarnCoins      EffectCause = "EARN_COINS"
	EffectStealCoins     EffectCause = "STEAL_COINS"
	EffectCardKilled     EffectCause = "CARD_KILLED"
	EffectRevivedCard    EffectCause = "CARD_REVIVED"
	EffectChangedCard    EffectCause = "CARD_CHANGED"
	EffectInvestment     EffectCause = "INVESTMENT"

	EffectAssassin     EffectCause = "ASSASSIN"
	EffectKamikaze     EffectCause = "KAMIKAZE"
	EffectTrillionaire EffectCause = "TRILLIONAIRE"
	EffectPolitical    EffectCause = "POLITICAL"
	EffectRebel        EffectCause = "REBEL"
	EffectClairvoyant  EffectCause = "CLAIRVOYANT"
	EffectGuardian     EffectCause = "GUARDIAN"
	EffectTrickster    EffectCause = "TRICKSTER"
	EffectGravedigger  EffectCause = "GRAVEDIGGER"
	EffectCroupier     EffectCause = "CROUPIER"
	EffectInvestor     EffectCause = "INVESTOR"
)
