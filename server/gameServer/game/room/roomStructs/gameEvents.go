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
	TypeUseCard          TypeGameEvents = "USE_CARD"
	TypeContest          TypeGameEvents = "CONTEST"
	TypeChooseCard       TypeGameEvents = "CHOOSE_CARD"
	TypeKamikazeDecision TypeGameEvents = "KAMIKAZE_DECISION"
	TypeTimeout          TypeGameEvents = "TIMEOUT"
)
