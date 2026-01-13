package roomStructs

type CardName string

type Card struct {
	Name      CardName `json:"name"`
	Index     int      `json:"index"`
	Protected bool     `json:"protected"`
	Dead      bool     `json:"dead"`
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
