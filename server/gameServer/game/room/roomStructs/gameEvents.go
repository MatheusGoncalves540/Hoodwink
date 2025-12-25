package roomStructs

import "time"

type GameEventType string
type EffectCause string

// ===== EVENTO (janela de decisão / input) =====
type PendingEvent struct {
	PlayerID  string
	Type      GameEventType
	ExpiresAt time.Time
	Payload   any
}

// Eventos = INPUT / JANELAS
const (
	EventWaitingFirstAction GameEventType = "WAITING_FIRST_ACTION"
	EventCardPlayedAssassin GameEventType = "CARD_PLAYED_ASSASSIN"
	EventCardKilled         GameEventType = "CARD_KILLED"
	EventShowingMessage     GameEventType = "SHOWING_MESSAGE"
)

// ===== EFEITO (precisa ser resolvido) =====
type Effect struct {
	Cause        EffectCause
	SourcePlayer string
	Payload      any
}

// Tipos de efeitos reais do jogo
const (
	EffectAssassin       EffectCause = "ASSASSIN"
	EffectContestPenalty EffectCause = "CONTEST_PENALTY"
	EffectKamikaze       EffectCause = "KAMIKAZE"
	EffectGainCoins      EffectCause = "GAIN_COINS"
)
