package rules

import "testing"

func TestClassicRulesMatchPDF(t *testing.T) {
	gameRules, err := LoadRules("modes/ClassicRules.yaml")
	if err != nil {
		t.Fatalf("erro ao carregar ClassicRules.yaml: %v", err)
	}

	if gameRules.Generals.MinTax == nil || *gameRules.Generals.MinTax != -5 {
		t.Fatalf("minTax = %v, quer -5", gameRules.Generals.MinTax)
	}
	if gameRules.Generals.MaxTax == nil || *gameRules.Generals.MaxTax != 5 {
		t.Fatalf("maxTax = %v, quer 5", gameRules.Generals.MaxTax)
	}
	if gameRules.Generals.MaxCoins == nil || *gameRules.Generals.MaxCoins != 20 {
		t.Fatalf("maxCoins = %v, quer 20", gameRules.Generals.MaxCoins)
	}
	if gameRules.Generals.RandomTaxesChanges == nil || *gameRules.Generals.RandomTaxesChanges != 0 {
		t.Fatalf("randomTaxesChanges = %v, quer 0", gameRules.Generals.RandomTaxesChanges)
	}
	if gameRules.Generals.MaxPlayers == nil || *gameRules.Generals.MaxPlayers != 10 {
		t.Fatalf("maxPlayers = %v, quer 10", gameRules.Generals.MaxPlayers)
	}
	if gameRules.Generals.BasicTakeCoins == nil || *gameRules.Generals.BasicTakeCoins != 1 {
		t.Fatalf("basicTakeCoins = %v, quer 1", gameRules.Generals.BasicTakeCoins)
	}
	if gameRules.Generals.BasicKillCardCost == nil || *gameRules.Generals.BasicKillCardCost != 6 {
		t.Fatalf("basicKillCardCost = %v, quer 6", gameRules.Generals.BasicKillCardCost)
	}
	if gameRules.Generals.BasicSwapMinCoins == nil || *gameRules.Generals.BasicSwapMinCoins != 6 {
		t.Fatalf("basicSwapMinCoins = %v, quer 6", gameRules.Generals.BasicSwapMinCoins)
	}

	assertCardValue(t, gameRules, "ASSASSIN", 3, 2, nil)
	assertCardValue(t, gameRules, "GRAVEDIGGER", 5, 3, nil)
	assertCardValue(t, gameRules, "CROUPIER", 6, 4, nil)
	assertCardValue(t, gameRules, "CLAIRVOYANT", 4, 2, nil)
	assertCardValue(t, gameRules, "TRILLIONAIRE", 3, 1, nil)
	assertCardValue(t, gameRules, "TRICKSTER", 2, 2, intPtr(5))
	assertCardValue(t, gameRules, "INVESTOR", 4, 2, intPtr(6))
	assertCardValue(t, gameRules, "SELFISH", 5, 3, intPtr(8))

	guardian := gameRules.Cards["GUARDIAN"]
	if guardian.TaxMinimum == nil || *guardian.TaxMinimum != 4 {
		t.Fatalf("Guardiã: taxMinimum = %v, quer 4", guardian.TaxMinimum)
	}
	if guardian.FutureProtectionTaxMinimum == nil || *guardian.FutureProtectionTaxMinimum != 5 {
		t.Fatalf("Guardiã: futureProtectionTaxMinimum = %v, quer 5", guardian.FutureProtectionTaxMinimum)
	}
	if guardian.FutureProtectionSurcharge == nil || *guardian.FutureProtectionSurcharge != 1 {
		t.Fatalf("Guardiã: futureProtectionSurcharge = %v, quer 1", guardian.FutureProtectionSurcharge)
	}
	if guardian.ClairvoyantBlockDiscount == nil || *guardian.ClairvoyantBlockDiscount != 1 {
		t.Fatalf("Guardiã: clairvoyantBlockDiscount = %v, quer 1", guardian.ClairvoyantBlockDiscount)
	}
	if guardian.ClairvoyantBlockTaxMinimum == nil || *guardian.ClairvoyantBlockTaxMinimum != 2 {
		t.Fatalf("Guardiã: clairvoyantBlockTaxMinimum = %v, quer 2", guardian.ClairvoyantBlockTaxMinimum)
	}

	investor := gameRules.Cards["INVESTOR"]
	if investor.GracePeriod == nil || *investor.GracePeriod != 6 {
		t.Fatalf("Investidor: gracePeriod = %v, quer 6", investor.GracePeriod)
	}

	selfish := gameRules.Cards["SELFISH"]
	if selfish.FixedValue == nil || *selfish.FixedValue != 1 {
		t.Fatalf("Egoísta: fixedValue = %v, quer 1", selfish.FixedValue)
	}
	if selfish.CoinDenialThreshold == nil || *selfish.CoinDenialThreshold != 2 {
		t.Fatalf("Egoísta: coinDenialThreshold = %v, quer 2", selfish.CoinDenialThreshold)
	}

	political := gameRules.Cards["POLITICAL"]
	if political.Value == nil || *political.Value != 2 || political.FixedValue == nil || *political.FixedValue != 1 {
		t.Fatalf("Político: value/fixedValue = %v/%v, quer 2/1", political.Value, political.FixedValue)
	}
	if political.AffectedByTaxes == nil || *political.AffectedByTaxes {
		t.Fatalf("Político não deveria ser afetado por taxas")
	}

	rebel := gameRules.Cards["REBEL"]
	if rebel.Value == nil || *rebel.Value != 2 || rebel.FixedValue == nil || *rebel.FixedValue != 1 {
		t.Fatalf("Rebelde: value/fixedValue = %v/%v, quer 2/1", rebel.Value, rebel.FixedValue)
	}
	if rebel.AffectedByTaxes == nil || *rebel.AffectedByTaxes {
		t.Fatalf("Rebelde não deveria ser afetado por taxas")
	}

	kamikaze := gameRules.Cards["KAMIKAZE"]
	if kamikaze.CanKillSelf == nil || !*kamikaze.CanKillSelf {
		t.Fatal("Kamikaze deveria permitir matar a si mesmo")
	}
	if kamikaze.AffectedByTaxes == nil || *kamikaze.AffectedByTaxes {
		t.Fatal("Kamikaze não deveria ser afetado por taxas")
	}
}

func assertCardValue(t *testing.T, gameRules *GameRules, card string, value, min int, max *int) {
	t.Helper()

	cardRules, ok := gameRules.Cards[card]
	if !ok {
		t.Fatalf("carta %s não encontrada", card)
	}
	if cardRules.Value == nil || *cardRules.Value != value {
		t.Fatalf("%s: value = %v, quer %d", card, cardRules.Value, value)
	}
	if cardRules.TaxMinimum == nil || *cardRules.TaxMinimum != min {
		t.Fatalf("%s: taxMinimum = %v, quer %d", card, cardRules.TaxMinimum, min)
	}
	if max == nil {
		if cardRules.TaxMaximum != nil {
			t.Fatalf("%s: taxMaximum = %v, quer nil", card, *cardRules.TaxMaximum)
		}
		return
	}
	if cardRules.TaxMaximum == nil || *cardRules.TaxMaximum != *max {
		t.Fatalf("%s: taxMaximum = %v, quer %d", card, cardRules.TaxMaximum, *max)
	}
}

func intPtr(v int) *int {
	return &v
}
