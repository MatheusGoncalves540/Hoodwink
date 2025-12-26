package roomStructs

type TypePlayerPlays string

type PlayerPlay struct {
	Type     TypePlayerPlays
	PlayerId string
	Payload  any
}

const (
	PlayAssassinCard TypePlayerPlays = "PLAY_ASSASSIN_CARD"
)
