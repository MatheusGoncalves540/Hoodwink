package rooms

import (
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
)

func loadClassicRegistry(t *testing.T) *rules.Registry {
	t.Helper()

	gameRules, err := rules.LoadRules("../../rules/modes/ClassicRules.yaml")
	if err != nil {
		t.Fatalf("erro ao carregar ClassicRules.yaml: %v", err)
	}
	registry := rules.NewRegistry()
	registry.Register(string(structs.ClassicRules), gameRules)
	return registry
}

func newClassicRoom() *Room {
	return &Room{
		Rules:             structs.ClassicRules,
		Tax:               0,
		DoubledCardValues: map[structs.TypePlayerPlays]structs.DoubledCardValues{},
		Players:           map[string]*players.Player{},
	}
}

func TestBasicActionCosts(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()

	killCases := []struct {
		tax  int
		want int
	}{
		{-5, 3},
		{-4, 3},
		{-3, 3},
		{-2, 4},
		{-1, 5},
		{0, 6},
		{1, 6},
		{3, 7},
		{5, 8},
	}
	for _, tc := range killCases {
		room.Tax = tc.tax
		killCost, err := room.GetBasicKillCardCost(registry)
		if err != nil {
			t.Fatal(err)
		}
		if killCost != tc.want {
			t.Fatalf("custo de matar com tax %+d = %d, quer %d", tc.tax, killCost, tc.want)
		}
	}

	room.Tax = -5
	amount, _ := room.GetBasicTakeCoinsAmount(registry)
	if amount != 1 {
		t.Fatalf("moedas ao coletar com tax -5 = %d, quer 1", amount)
	}

	room.Tax = 5
	amount, _ = room.GetBasicTakeCoinsAmount(registry)
	if amount != 6 {
		t.Fatalf("moedas ao coletar com tax +5 = %d, quer 6", amount)
	}

	if got := room.GetBasicSwapCost(7, false); got != 4 {
		t.Fatalf("custo de troca de 1 carta com 7 moedas = %d, quer 4", got)
	}
	if got := room.GetBasicSwapCost(20, true); got != 20 {
		t.Fatalf("custo de troca de 2 cartas = %d, quer 20", got)
	}
}

func TestCardValueRespectsTaxLimits(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()

	cases := []struct {
		tax  int
		role structs.TypePlayerPlays
		want int
	}{
		{-5, structs.PlayAssassinCard, 2},
		{5, structs.PlayAssassinCard, 8},
		{-5, structs.PlayGravediggerCard, 3},
		{0, structs.PlayGravediggerCard, 5},
		{5, structs.PlayGravediggerCard, 10},
		{-5, structs.PlayCroupierCard, 4},
		{0, structs.PlayCroupierCard, 6},
		{5, structs.PlayCroupierCard, 11},
		{-5, structs.PlaySelfishCard, 3},
		{5, structs.PlaySelfishCard, 8},
		{-5, structs.PlayClairvoyantCard, 2},
		{0, structs.PlayClairvoyantCard, 4},
		{-5, structs.PlayTrillionaireCard, 1},
		{0, structs.PlayTrillionaireCard, 3},
		{5, structs.PlayTricksterCard, 5},
		{-5, structs.PlayInvestorCard, 2},
		{5, structs.PlayInvestorCard, 6},
	}

	for _, tc := range cases {
		room.Tax = tc.tax
		value, err := room.GetCardValue(registry, tc.role, 0)
		if err != nil {
			t.Fatalf("%s com tax %+d: %v", tc.role, tc.tax, err)
		}
		if value != tc.want {
			t.Fatalf("%s com tax %+d = %d, quer %d", tc.role, tc.tax, value, tc.want)
		}
	}
}

func TestTaxImmuneRoles(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()

	immune := []structs.TypePlayerPlays{
		structs.PlayPoliticalCard,
		structs.PlayRebelCard,
	}

	for _, role := range immune {
		for _, tax := range []int{-5, 0, 5} {
			room.Tax = tax
			value, err := room.GetCardValue(registry, role, 0)
			if err != nil {
				t.Fatalf("%s com tax %+d: %v", role, tax, err)
			}
			if value != 2 {
				t.Fatalf("%s com tax %+d = %d, quer 2 (imune à taxa)", role, tax, value)
			}
		}
	}
}

func TestInvestorGracePeriodIsSixTurns(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()

	cardRules, err := room.GetCardRules(registry, string(structs.PlayInvestorCard))
	if err != nil {
		t.Fatal(err)
	}
	if cardRules.GracePeriod == nil {
		t.Fatal("gracePeriod não configurado para o Investidor")
	}
	if *cardRules.GracePeriod != 6 {
		t.Fatalf("gracePeriod = %d, quer 6", *cardRules.GracePeriod)
	}
}

func TestGuardianPrices(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()

	cases := []struct {
		tax    int
		death  int
		sight  int
		future int
	}{
		{-5, 4, 2, 5},
		{0, 5, 4, 6},
		{5, 10, 9, 11},
	}

	for _, tc := range cases {
		room.Tax = tc.tax

		if got, _ := room.GetGuardianValue(registry, "death"); got != tc.death {
			t.Fatalf("Guardiã imediata com tax %+d = %d, quer %d", tc.tax, got, tc.death)
		}
		if got, _ := room.GetGuardianValue(registry, "sight"); got != tc.sight {
			t.Fatalf("Guardiã contra Clairvoyant com tax %+d = %d, quer %d", tc.tax, got, tc.sight)
		}
		if got, _ := room.GetGuardianValue(registry, "future"); got != tc.future {
			t.Fatalf("Guardiã futura com tax %+d = %d, quer %d", tc.tax, got, tc.future)
		}
	}
}
