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
