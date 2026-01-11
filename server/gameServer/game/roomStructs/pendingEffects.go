package roomStructs

type EffectCause string

// ===== EFEITO (precisa ser executado) =====
type Effect struct {
	Cause        EffectCause
	SourcePlayer string
	Payload      any
}

// Tipos de efeitos
const (
	EffectAssassin       EffectCause = "ASSASSIN"
	EffectContestPenalty EffectCause = "CONTEST_PENALTY"
	EffectKamikaze       EffectCause = "KAMIKAZE"
	EffectGainCoins      EffectCause = "GAIN_COINS"
)
