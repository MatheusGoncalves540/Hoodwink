package rooms

import (
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
)

func TestChangeCardClearsRevealState(t *testing.T) {
	player := &players.Player{
		Id:    "p1",
		Cards: []structs.Card{{Name: structs.CardAssassin, Index: 0, RevealedToAll: true}},
	}
	room := &Room{Deck: []structs.CardName{structs.CardRebel}}

	if err := room.ChangeCard(player, 0); err != nil {
		t.Fatalf("erro ao trocar carta: %v", err)
	}
	if player.Cards[0].RevealedToAll {
		t.Fatal("carta trocada continuou revelada para todos")
	}
	if player.Cards[0].Name != structs.CardRebel {
		t.Fatalf("nome da carta trocada = %q, quer %q", player.Cards[0].Name, structs.CardRebel)
	}
}
