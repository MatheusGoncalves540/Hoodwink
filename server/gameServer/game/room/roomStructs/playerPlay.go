package roomStructs

import "encoding/json"

type TypePlayerPlays string

type PlayerPlay struct {
	Type     TypePlayerPlays
	PlayerId string
	Payload  any
}

const (
	PlayAssassinCard TypePlayerPlays = "PLAY_ASSASSIN_CARD"
)

//--//--//--//--//--//--//--//--//--//

type PlayerPlayPayload struct {
	Type     TypePlayerPlays `json:"type"`
	PlayerID string          `json:"playerId"`
	Payload  json.RawMessage `json:"payload"`
}
