package structs

type KillCardPayload struct {
	Cause           string  `json:"cause"`
	TargetPlayer    *string `json:"targetPlayer"`
	TargetCardIndex *int    `json:"targetCardIndex"`
}

type ContestPayload struct {
	ContestedPlayer string `json:"contestedPlayer"`
	ContestedCard   string `json:"contestedCard"`
}

type ContestPenaltyPayload struct {
	ContestedPlayer string `json:"contestedPlayer"`
	ContestedCard   string `json:"contestedCard"`
	HasCard         bool   `json:"hasCard"`
	TargetCardIndex *int   `json:"targetCardIndex,omitempty"`
}

type AssassinPayload struct {
	TargetPlayer    *string `json:"targetPlayer"`
	TargetCardIndex *int    `json:"targetCardIndex"`
}

type KamikazePayload struct {
	TargetPlayer        *string `json:"targetPlayer"`
	TargetCardIndex     *int    `json:"targetCardIndex"`
	TargetAllyCardIndex *int    `json:"targetAllyCardIndex,omitempty"`
	KilledHimSelf       bool    `json:"killedHimself,omitempty"`
}
