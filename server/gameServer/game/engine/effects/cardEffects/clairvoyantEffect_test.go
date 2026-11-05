package cardEffects

import (
	"context"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func newClairvoyantRoom() (*rooms.Room, *players.Player) {
	target := &players.Player{
		Id:    "p1",
		Cards: []structs.Card{{Name: structs.CardAssassin, Index: 0}},
	}
	room := &rooms.Room{Players: map[string]*players.Player{"p1": target}}
	return room, target
}

func TestClairvoyantShowToAllMarksCardRevealed(t *testing.T) {
	room, target := newClairvoyantRoom()
	effect := structs.NewEffect(structs.EffectClairvoyant, "p2", structs.NewClairvoyantPayload("p1", 0, true, nil))

	ClairvoyantEffect(context.Background(), nil, nil, room, effect)

	if !target.Cards[0].RevealedToAll {
		t.Fatal("carta revelada para todos não foi marcada como revelada")
	}
}

func TestClairvoyantPrivateDoesNotMarkCardRevealed(t *testing.T) {
	room, target := newClairvoyantRoom()
	effect := structs.NewEffect(structs.EffectClairvoyant, "p2", structs.NewClairvoyantPayload("p1", 0, false, nil))

	ClairvoyantEffect(context.Background(), nil, nil, room, effect)

	if target.Cards[0].RevealedToAll {
		t.Fatal("carta revelada só para o jogador foi marcada como revelada para todos")
	}
}
