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
	EffectKamikaze       EffectCause = "KAMIKAZE"
	EffectContestPenalty EffectCause = "CONTEST_PENALTY"
	EffectGainCoins      EffectCause = "GAIN_COINS"
)
