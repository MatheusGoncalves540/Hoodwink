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

//

type AssassinPayload struct {
	TargetPlayer string `json:"targetPlayer"`
	TargetCard   int    `json:"targetCard"`
}
