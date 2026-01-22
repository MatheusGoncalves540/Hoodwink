package structs

import "encoding/json"

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

// Tipos de efeitos
const (
	EffectAssassin       EffectCause = "ASSASSIN"
	EffectKamikaze       EffectCause = "KAMIKAZE"
	EffectTrillionaire   EffectCause = "TRILLIONAIRE"
	EffectContest        EffectCause = "CONTEST"
	EffectContestPenalty EffectCause = "CONTEST_PENALTY"
	EffectGreed          EffectCause = "GREED"
	EffectEarnCoins      EffectCause = "EARN_COINS"
)
