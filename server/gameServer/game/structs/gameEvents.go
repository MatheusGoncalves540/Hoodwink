package structs

import (
	"time"
)

type GameEventType string

// ===== EVENTO (janela de decisão / input) =====
type GameEvent struct {
	PlayerID  string
	Type      GameEventType
	ExpiresAt time.Time
	Payload   any
}

// NewGameEvent centralizes creation of a GameEvent pointer.
func NewGameEvent(playerID string, eventType GameEventType, expiresAt time.Time, payload any) *GameEvent {
	return &GameEvent{
		PlayerID:  playerID,
		Type:      eventType,
		ExpiresAt: expiresAt,
		Payload:   payload,
	}
}

// Eventos = INPUT / JANELAS
const (
	EventWaitingFirstAction     GameEventType = "WAITING_FIRST_ACTION"
	EventCardKilled             GameEventType = "CARD_KILLED"
	EventContest                GameEventType = "CONTEST"
	EventContestPenalty         GameEventType = "CONTEST_PENALTY"
	EventIncrementTaxes         GameEventType = "INCREMENT_TAXES"
	EventDecrementTaxes         GameEventType = "DECREMENT_TAXES"
	EventEarnCoins              GameEventType = "EARN_COINS"
	EventCardPlayedAssassin     GameEventType = "CARD_PLAYED_ASSASSIN"
	EventCardPlayedKamikaze     GameEventType = "CARD_PLAYED_KAMIKAZE"
	EventCardPlayedTrillionaire GameEventType = "CARD_PLAYED_TRILLIONAIRE"
	EventCardPlayedPolitical    GameEventType = "CARD_PLAYED_POLITICAL"
	EventCardPlayedRebel        GameEventType = "CARD_PLAYED_REBEL"
	EventCardPlayedClairvoyant  GameEventType = "CARD_PLAYED_CLAIRVOYANT"
	EventRevealedCard           GameEventType = "REVEALED_CARD"
)
