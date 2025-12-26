package roomStructs

type AssassinPayload struct {
	TargetPlayer *string `json:"targetPlayer"`
	TargetCard   *int    `json:"targetCard"`
}
