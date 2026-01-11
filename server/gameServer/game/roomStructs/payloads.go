package roomStructs

type KillCardPayload struct {
	TargetPlayer    *string `json:"targetPlayer"`
	TargetCardIndex *int    `json:"targetCardIndex"`
}

type AssassinPayload struct {
	TargetPlayer    *string `json:"targetPlayer"`
	TargetCardIndex *int    `json:"targetCardIndex"`
}
