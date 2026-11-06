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

func TestClairvoyantRevealDoesNotPersistCardState(t *testing.T) {
	room, target := newClairvoyantRoom()
	effect := structs.NewEffect(structs.EffectClairvoyant, "p2", structs.NewClairvoyantPayload("p1", 0, true, nil))

	ClairvoyantEffect(context.Background(), nil, nil, room, effect)

	if target.Cards[0].RevealedToAll {
		t.Fatal("carta revelada pelo Clarividente ficou marcada como revelada permanentemente")
	}
	if len(room.PendingPresentationEvents) != 1 {
		t.Fatalf("eventos de apresentação = %d, quer 1", len(room.PendingPresentationEvents))
	}
}

func TestClairvoyantPrivateRevealUsesConfidencialPayload(t *testing.T) {
	room, target := newClairvoyantRoom()
	effect := structs.NewEffect(structs.EffectClairvoyant, "p2", structs.NewClairvoyantPayload("p1", 0, false, nil))

	ClairvoyantEffect(context.Background(), nil, nil, room, effect)

	if target.Cards[0].RevealedToAll {
		t.Fatal("carta revelada só para o jogador foi marcada como revelada para todos")
	}
	if len(room.PendingPresentationEvents) != 1 {
		t.Fatalf("eventos de apresentação = %d, quer 1", len(room.PendingPresentationEvents))
	}
	if len(room.PendingPresentationEvents[0].ConfidencialPlayerIds) != 1 || room.PendingPresentationEvents[0].ConfidencialPlayerIds[0] != "p2" {
		t.Fatalf("payload confidencial não foi direcionado apenas para quem usou o Clarividente")
	}
}
