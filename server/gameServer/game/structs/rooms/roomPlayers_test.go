package rooms

import (
	"fmt"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
)

func TestClassicRulesAllowsTenPlayers(t *testing.T) {
	registry := loadClassicRegistry(t)
	room := newClassicRoom()

	generalRules, err := room.GetGeneralRules(registry)
	if err != nil {
		t.Fatal(err)
	}
	if generalRules.MaxPlayers == nil {
		t.Fatal("maxPlayers não configurado no ClassicRules.yaml")
	}
	if *generalRules.MaxPlayers != 10 {
		t.Fatalf("maxPlayers = %d, quer 10", *generalRules.MaxPlayers)
	}
}

func TestRoomAcceptsTenthPlayerAndRejectsEleventh(t *testing.T) {
	room := &Room{
		MaxPlayers: 10,
		Players:    map[string]*players.Player{},
	}

	for i := 1; i <= 10; i++ {
		if room.IsFull() {
			t.Fatalf("sala recusou o %dº jogador (máximo %d)", i, room.MaxPlayers)
		}
		id := fmt.Sprintf("p%d", i)
		room.Players[id] = &players.Player{Id: id}
	}

	if !room.IsFull() {
		t.Fatalf("sala com 10 jogadores deveria recusar o 11º")
	}
}
