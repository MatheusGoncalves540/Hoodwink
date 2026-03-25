package structs

import "math/rand/v2"

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
	CardAssassin     CardName = "ASSASSIN"
	CardKamikaze     CardName = "KAMIKAZE"
	CardPolitical    CardName = "POLITICAL"
	CardRebel        CardName = "REBEL"
	CardTrillionaire CardName = "TRILLIONAIRE"
	CardClairvoyant  CardName = "CLAIRVOYANT"
	CardGuardian     CardName = "GUARDIAN"
	CardTrickster    CardName = "TRICKSTER"
	CardGravedigger  CardName = "GRAVEDIGGER"
	CardCroupier     CardName = "CROUPIER"
	CardInvestor     CardName = "INVESTOR"
	CardSelfish      CardName = "SELFISH"
)

var AllCards = []CardName{
	CardAssassin,
	CardKamikaze,
	CardPolitical,
	CardRebel,
	CardTrillionaire,
	CardClairvoyant,
	CardGuardian,
	CardTrickster,
	CardGravedigger,
	CardCroupier,
	CardInvestor,
	CardSelfish,
}

func RandomCard() CardName {
	return AllCards[rand.IntN(len(AllCards))]
}
