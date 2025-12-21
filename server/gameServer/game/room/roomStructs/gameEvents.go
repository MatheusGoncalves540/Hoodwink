package roomStructs

type TypeGameEvents string

// Estrutura Dos eventos do jogo
type PlayerPlay struct {
	PlayerID string
	Type     string
	Payload  any
}

// Tipos possiveis de Eventos
const (
	TypeUseCard          TypeGameEvents = "USE_CARD"
	TypeContest          TypeGameEvents = "CONTEST"
	TypeChooseCard       TypeGameEvents = "CHOOSE_CARD"
	TypeKamikazeDecision TypeGameEvents = "KAMIKAZE_DECISION"
	TypeTimeout          TypeGameEvents = "TIMEOUT"
)
