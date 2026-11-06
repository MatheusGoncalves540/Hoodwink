package effects

import (
	"context"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func TestReviveCardEffectRevivesTheSameChosenCard(t *testing.T) {
	player := &players.Player{
		Id:    "p1",
		Alive: true,
		Cards: []structs.Card{
			{Name: structs.CardAssassin, Index: 0, Dead: true},
			{Name: structs.CardGuardian, Index: 1},
		},
	}
	room := &rooms.Room{Players: map[string]*players.Player{"p1": player}}

	targetPlayer := player.Id
	targetCardIndex := 0
	effect := structs.NewEffect(
		structs.EffectGravedigger,
		player.Id,
		structs.NewReviveCardPayload(string(structs.EffectGravedigger), &targetPlayer, &targetCardIndex),
	)

	if err := ReviveCardEffect(context.Background(), nil, nil, room, effect); err != nil {
		t.Fatal(err)
	}

	if player.Cards[0].Dead {
		t.Fatal("a carta escolhida deveria estar viva após o Coveiro")
	}
	if player.Cards[0].Name != structs.CardAssassin {
		t.Fatalf("carta revivida = %s, quer a mesma carta morta (ASSASSIN)", player.Cards[0].Name)
	}
	if player.Cards[1].Name != structs.CardGuardian || player.Cards[1].Dead {
		t.Fatal("o Coveiro não deveria alterar a outra carta do jogador")
	}
}
