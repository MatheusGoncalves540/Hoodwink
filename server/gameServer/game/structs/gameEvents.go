package structs

import "time"

type GameEventType string

// ===== EVENTO (janela de decisão / input) =====
type GameEvent struct {
	PlayerID  string
	Type      GameEventType
	ExpiresAt time.Time
	Payload   any
}

// Eventos = INPUT / JANELAS
const (
	EventWaitingFirstAction GameEventType = "WAITING_FIRST_ACTION"
	EventCardPlayedAssassin GameEventType = "CARD_PLAYED_ASSASSIN"
	EventCardPlayedKamikaze GameEventType = "CARD_PLAYED_KAMIKAZE"
	EventCardKilled         GameEventType = "CARD_KILLED"
	EventShowingMessage     GameEventType = "SHOWING_MESSAGE"
)
