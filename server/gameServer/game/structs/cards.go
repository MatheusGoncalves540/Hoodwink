package structs

type CardName string

type Card struct {
	Name      CardName `json:"name"`
	Index     int      `json:"index"`
	Protected bool     `json:"protected"`
	Dead      bool     `json:"dead"`
}

type DoubledCardValues struct {
	TimesValueDoubled   int  `json:"timesValueDoubled"`
	RoundsUntilDecrease int  `json:"roundsUntilDecrease"`
	UsedThisTurn        bool `json:"usedThisTurn"`
}

type PublicCardForUpdates struct {
	Index     int  `json:"index"`
	Protected bool `json:"protected"`
	Dead      bool `json:"dead"`
}

const (
	CardAssassin CardName = "ASSASSIN"
	CardKamikaze CardName = "KAMIKAZE"
)
