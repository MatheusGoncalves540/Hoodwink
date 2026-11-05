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
	if amount != 0 {
		t.Fatalf("moedas ao coletar com tax -5 = %d, quer 0", amount)
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

	room.Tax = -5
	value, err := room.GetCardValue(registry, structs.PlayAssassinCard, 0)
	if err != nil {
		t.Fatal(err)
	}
	if value != 2 {
		t.Fatalf("Assassino com tax -5 = %d, quer 2", value)
	}

	room.Tax = 5
	value, _ = room.GetCardValue(registry, structs.PlayAssassinCard, 0)
	if value != 8 {
		t.Fatalf("Assassino com tax +5 = %d, quer 8", value)
	}

	value, _ = room.GetCardValue(registry, structs.PlayTricksterCard, 0)
	if value != 5 {
		t.Fatalf("Trapaceiro com tax +5 = %d, quer 5", value)
	}

	value, _ = room.GetCardValue(registry, structs.PlayInvestorCard, 0)
	if value != 6 {
		t.Fatalf("Investidor com tax +5 = %d, quer 6", value)
	}

	room.Tax = -5
	value, _ = room.GetCardValue(registry, structs.PlayInvestorCard, 0)
	if value != 2 {
		t.Fatalf("Investidor com tax -5 = %d, quer 2", value)
	}
}

func TestGuardianPrices(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()

	death, _ := room.GetGuardianValue(registry, "death")
	if death != 5 {
		t.Fatalf("Guardiã imediata = %d, quer 5", death)
	}

	sight, _ := room.GetGuardianValue(registry, "sight")
	if sight != 4 {
		t.Fatalf("Guardiã contra Clairvoyant = %d, quer 4", sight)
	}

	future, _ := room.GetGuardianValue(registry, "future")
	if future != 6 {
		t.Fatalf("Guardiã futura = %d, quer 6", future)
	}

	room.Tax = -5
	future, _ = room.GetGuardianValue(registry, "future")
	if future != 4 {
		t.Fatalf("Guardiã futura com tax -5 = %d, quer 4", future)
	}

	death, _ = room.GetGuardianValue(registry, "death")
	if death != 2 {
		t.Fatalf("Guardiã imediata com tax -5 = %d, quer 2", death)
	}

	room.Tax = 5
	death, _ = room.GetGuardianValue(registry, "death")
	if death != 10 {
		t.Fatalf("Guardiã imediata com tax +5 = %d, quer 10", death)
	}
	sight, _ = room.GetGuardianValue(registry, "sight")
	if sight != 9 {
		t.Fatalf("Guardiã contra Clairvoyant com tax +5 = %d, quer 9", sight)
	}
}
