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
	EventWaitingFirstAction     GameEventType = "WAITING_FIRST_ACTION"
	EventCardKilled             GameEventType = "CARD_KILLED"
	EventContest                GameEventType = "CONTEST"
	EventContestPenalty         GameEventType = "CONTEST_PENALTY"
	EventCardPlayedAssassin     GameEventType = "CARD_PLAYED_ASSASSIN"
	EventCardPlayedKamikaze     GameEventType = "CARD_PLAYED_KAMIKAZE"
	EventCardPlayedTrillionaire GameEventType = "CARD_PLAYED_TRILLIONAIRE"
	EventEarnCoins              GameEventType = "EARN_COINS"
)
