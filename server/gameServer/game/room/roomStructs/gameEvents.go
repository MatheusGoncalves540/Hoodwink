package roomStructs

import "time"

type TypeGameEvents string

// Estrutura Dos eventos do jogo
type PendingEvent struct {
	PlayerID  string
	Type      TypeGameEvents
	ExpiresAt time.Time
	Payload   any
}

// Tipos possiveis de Eventos
const (
	TypeDisplayingMessage TypeGameEvents = "DISPLAYING_MESSAGE"
	TypeUseCard           TypeGameEvents = "USE_CARD"
	TypePickupCoin        TypeGameEvents = "PICKUP_COIN"
	TypePass              TypeGameEvents = "PASS"
	TypeContest           TypeGameEvents = "CONTEST"
)
