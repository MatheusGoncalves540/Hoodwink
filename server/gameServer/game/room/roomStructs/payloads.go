package roomStructs

type KillCardPayload struct {
	TargetPlayer *string `json:"targetPlayer"`
	TargetCard   *int    `json:"targetCard"`
}

type AssassinPayload struct {
	TargetPlayer *string `json:"targetPlayer"`
	TargetCard   *int    `json:"targetCard"`
}
