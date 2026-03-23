package structs

type KillCardPayload struct {
	Cause           string  `json:"cause"`
	TargetPlayer    *string `json:"targetPlayer"`
	TargetCardIndex *int    `json:"targetCardIndex"`
}

func NewKillCardPayload(cause string, targetPlayer *string, targetCardIndex *int) KillCardPayload {
	return KillCardPayload{
		Cause:           cause,
		TargetPlayer:    targetPlayer,
		TargetCardIndex: targetCardIndex,
	}
}

type StealCoinsPayload struct {
	Cause        string  `json:"cause"`
	LostCoins    int     `json:"lostCoins,omitempty"`
	TargetPlayer *string `json:"targetPlayer,omitempty"`
}

func NewStealCoinsPayload(cause string, lostCoins int, targetPlayer *string) StealCoinsPayload {
	return StealCoinsPayload{
		Cause:        cause,
		LostCoins:    lostCoins,
		TargetPlayer: targetPlayer,
	}
}

type EarnCoinsPayload struct {
	Cause       string `json:"cause"`
	EarnedCoins int    `json:"earnedCoins,omitempty"`
	BreakLimit  *bool  `json:"breakLimit,omitempty"`
}

func NewEarnCoinsPayload(cause string, earnedCoins int, breakLimit *bool) EarnCoinsPayload {
	return EarnCoinsPayload{
		Cause:       cause,
		EarnedCoins: earnedCoins,
		BreakLimit:  breakLimit,
	}
}

type ContestPayload struct {
	ContestedPlayer string `json:"contestedPlayer"`
	ContestedCard   string `json:"contestedCard"`
}

func NewContestPayload(contestedPlayer, contestedCard string) ContestPayload {
	return ContestPayload{
		ContestedPlayer: contestedPlayer,
		ContestedCard:   contestedCard,
	}
}

type ContestPenaltyPayload struct {
	ContestedPlayer string `json:"contestedPlayer"`
	ContestedCard   string `json:"contestedCard"`
	HasCard         bool   `json:"hasCard"`
	TargetCardIndex *int   `json:"targetCardIndex,omitempty"`
}

func NewContestPenaltyPayload(contestedPlayer, contestedCard string, hasCard bool, targetCardIndex *int) ContestPenaltyPayload {
	return ContestPenaltyPayload{
		ContestedPlayer: contestedPlayer,
		ContestedCard:   contestedCard,
		HasCard:         hasCard,
		TargetCardIndex: targetCardIndex,
	}
}

type AssassinPayload struct {
	TargetPlayer    *string `json:"targetPlayer"`
	TargetCardIndex *int    `json:"targetCardIndex"`
}

func NewAssassinPayload(targetPlayer *string, targetCardIndex *int) AssassinPayload {
	return AssassinPayload{
		TargetPlayer:    targetPlayer,
		TargetCardIndex: targetCardIndex,
	}
}

type KamikazePayload struct {
	TargetPlayer        *string `json:"targetPlayer"`
	TargetCardIndex     *int    `json:"targetCardIndex"`
	TargetAllyCardIndex *int    `json:"targetAllyCardIndex,omitempty"`
	KilledHimSelf       bool    `json:"killedHimself,omitempty"`
}

func NewKamikazePayload(targetPlayer *string, targetCardIndex *int, targetAllyCardIndex *int, killedHimSelf bool) KamikazePayload {
	return KamikazePayload{
		TargetPlayer:        targetPlayer,
		TargetCardIndex:     targetCardIndex,
		TargetAllyCardIndex: targetAllyCardIndex,
		KilledHimSelf:       killedHimSelf,
	}
}

type TrillionairePayload struct {
	EarnedCoins int `json:"earnedCoins,omitempty"`
}

func NewTrillionairePayload(earnedCoins int, breakLimit *bool) TrillionairePayload {
	return TrillionairePayload{
		EarnedCoins: earnedCoins,
	}
}

type PoliticalPayload struct {
	TaxIncrease int `json:"taxIncrease,omitempty"`
}

func NewPoliticalPayload(taxIncrease int) PoliticalPayload {
	return PoliticalPayload{
		TaxIncrease: taxIncrease,
	}
}

type RebelPayload struct {
	TaxReduction int `json:"taxReduction,omitempty"`
}

func NewRebelPayload(taxReduction int) RebelPayload {
	return RebelPayload{
		TaxReduction: taxReduction,
	}
}

type ClairvoyantPayload struct {
	TargetPlayer     string  `json:"targetPlayer"`
	TargetCardIndex  int     `json:"targetCardIndex"`
	ShowToAllPlayers bool    `json:"showToAllPlayers"`
	RevealedCard     *string `json:"revealedCard,omitempty"`
}

func NewClairvoyantPayload(targetPlayer string, targetCardIndex int, showToAllPlayers bool, revealedCard *string) ClairvoyantPayload {
	return ClairvoyantPayload{
		TargetPlayer:     targetPlayer,
		TargetCardIndex:  targetCardIndex,
		ShowToAllPlayers: showToAllPlayers,
		RevealedCard:     revealedCard,
	}
}

type GuardianPayload struct {
	TargetPlayer    string `json:"targetPlayer"`
	TargetCardIndex int    `json:"targetCardIndex"`
	ProtectedFrom   string `json:"protectedFrom,omitempty"`
}

func NewGuardianPayload(targetPlayer string, targetCardIndex int, protectedFrom string) GuardianPayload {
	return GuardianPayload{
		TargetPlayer:    targetPlayer,
		TargetCardIndex: targetCardIndex,
		ProtectedFrom:   protectedFrom,
	}
}

type TricksterPayload struct {
	TargetPlayer string `json:"targetPlayer"`
	StealedCoins int    `json:"stealedCoins,omitempty"`
}

func NewTricksterPayload(targetPlayer string, stealedCoins int) TricksterPayload {
	return TricksterPayload{
		TargetPlayer: targetPlayer,
		StealedCoins: stealedCoins,
	}
}
