package players

import (
	"reflect"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

func TestGetPublicPlayerForUpdatesCopiesTurnState(t *testing.T) {
	player := &Player{
		Id:                    "p1",
		Name:                  "Ana",
		MoveTime:              3,
		Cards:                 []structs.Card{{Name: structs.CardAssassin, Index: 0}},
		Coins:                 7,
		Investments:           []int{2, 4},
		PendingInvestmentCoin: true,
		Alive:                 true,
	}

	want := PublicPlayerForUpdates{
		Id:                    "p1",
		Name:                  "Ana",
		MoveTime:              3,
		Cards:                 []structs.PublicCardForUpdates{{Index: 0}},
		Coins:                 7,
		Investments:           []int{2, 4},
		PendingInvestmentCoin: true,
		Alive:                 true,
	}

	if got := player.GetPublicPlayerForUpdates(); !reflect.DeepEqual(got, want) {
		t.Fatalf("jogador público = %#v, quer %#v", got, want)
	}
}

func TestGetPublicPlayerForUpdatesSendsEmptyInvestmentsSlice(t *testing.T) {
	player := &Player{Id: "p1", Cards: []structs.Card{}}

	got := player.GetPublicPlayerForUpdates()
	if got.Investments == nil {
		t.Fatal("investments nulo no jogador público, quer []")
	}
	if len(got.Investments) != 0 {
		t.Fatalf("investments = %v, quer []", got.Investments)
	}
}

func TestGetPublicPlayerForUpdatesOnlyExposesKnownCards(t *testing.T) {
	player := &Player{
		Id: "p1",
		Cards: []structs.Card{
			{Name: structs.CardAssassin, Index: 0},
			{Name: structs.CardKamikaze, Index: 1, Dead: true},
			{Name: structs.CardGuardian, Index: 2, RevealedToAll: true},
			{Name: structs.CardTrickster, Index: 3, Protected: true},
		},
	}

	got := player.GetPublicPlayerForUpdates()

	if got.Cards[0].Name != nil {
		t.Fatalf("carta viva e oculta expôs o nome %q", *got.Cards[0].Name)
	}
	if got.Cards[1].Name == nil || *got.Cards[1].Name != structs.CardKamikaze {
		t.Fatalf("carta morta sem nome, quer %q", structs.CardKamikaze)
	}
	if got.Cards[2].Name == nil || *got.Cards[2].Name != structs.CardGuardian {
		t.Fatalf("carta revelada sem nome, quer %q", structs.CardGuardian)
	}
	if got.Cards[3].Name != nil {
		t.Fatalf("carta protegida e oculta expôs o nome %q", *got.Cards[3].Name)
	}
}

func TestGetPrivatePlayerForPublicUpdatesKeepsAllCardNames(t *testing.T) {
	player := &Player{
		Id:    "p1",
		Cards: []structs.Card{{Name: structs.CardAssassin, Index: 0}},
	}

	got := player.GetPrivatePlayerForPublicUpdates()
	if got.Cards[0].Name == nil || *got.Cards[0].Name != structs.CardAssassin {
		t.Fatalf("carta privada sem nome, quer %q", structs.CardAssassin)
	}
}

func TestGetPrivatePlayerForPublicUpdatesCopiesTurnState(t *testing.T) {
	player := &Player{
		Id:                    "p1",
		Name:                  "Ana",
		MoveTime:              3,
		Cards:                 []structs.Card{{Name: structs.CardAssassin, Index: 0}},
		Coins:                 7,
		Investments:           []int{2, 4},
		PendingInvestmentCoin: true,
		Alive:                 true,
	}

	got := player.GetPrivatePlayerForPublicUpdates()
	if got.MoveTime != player.MoveTime {
		t.Fatalf("moveTime = %d, quer %d", got.MoveTime, player.MoveTime)
	}
	if !reflect.DeepEqual(got.Investments, player.Investments) {
		t.Fatalf("investments = %v, quer %v", got.Investments, player.Investments)
	}
	if got.PendingInvestmentCoin != player.PendingInvestmentCoin {
		t.Fatalf("pendingInvestmentCoin = %t, quer %t", got.PendingInvestmentCoin, player.PendingInvestmentCoin)
	}
}
