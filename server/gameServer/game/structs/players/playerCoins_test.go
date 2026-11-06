package players

import "testing"

func TestAddCoinsClampsAtMaxAndSignalsCardLoss(t *testing.T) {
	player := &Player{Coins: 18}

	if !player.AddCoins(5, 20) {
		t.Fatal("esperava indicação de perda de carta ao ultrapassar o limite")
	}
	if player.Coins != 20 {
		t.Fatalf("moedas = %d, quer 20 (excedente descartado)", player.Coins)
	}
}

func TestAddCoinsDoesNotBreakLimitAtExactMax(t *testing.T) {
	player := &Player{Coins: 18}

	if player.AddCoins(2, 20) {
		t.Fatal("não deveria perder carta ao atingir exatamente o limite")
	}
	if player.Coins != 20 {
		t.Fatalf("moedas = %d, quer 20", player.Coins)
	}
}

func TestTrimInvestmentsDiscardsExcessWithoutReturningCoins(t *testing.T) {
	player := &Player{Investments: []int{3, 4, 5, 6}, PendingInvestmentCoin: true}

	discarded := player.TrimInvestments(2)
	if discarded != 2 {
		t.Fatalf("investimentos descartados = %d, quer 2", discarded)
	}
	if len(player.Investments) != 2 {
		t.Fatalf("investimentos restantes = %d, quer 2", len(player.Investments))
	}
	if player.PendingInvestmentCoin {
		t.Fatal("flag de investimento pendente deveria ser resetada")
	}
}
