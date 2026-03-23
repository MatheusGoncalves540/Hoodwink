package structs

import "encoding/json"

type TypePlayerPlays string

type PlayerPlay struct {
	Type     TypePlayerPlays
	PlayerId string
	Payload  any
}

const (
	PlayContest          TypePlayerPlays = "CONTEST"
	PlayContestPenalty   TypePlayerPlays = "CONTEST_PENALTY"
	PlayAssassinCard     TypePlayerPlays = "ASSASSIN"
	PlayKamikazeCard     TypePlayerPlays = "KAMIKAZE"
	PlayTrillionaireCard TypePlayerPlays = "TRILLIONAIRE"
	PlayPoliticalCard    TypePlayerPlays = "POLITICAL"
	PlayRebelCard        TypePlayerPlays = "REBEL"
	PlayClairvoyantCard  TypePlayerPlays = "CLAIRVOYANT"
	PlayGuardianCard     TypePlayerPlays = "GUARDIAN"
	PlayTricksterCard    TypePlayerPlays = "TRICKSTER"
)

//--//--//--//--//--//--//--//--//--//

type PlayerPlayDTO struct {
	Type     TypePlayerPlays `json:"type"`
	PlayerID string          `json:"playerId"`
	Payload  json.RawMessage `json:"payload,omitempty"`
}
