package rooms

import (
	"encoding/json"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
)

func TestGetPublicRoomForUpdatesIncludesNameAndMaxPlayers(t *testing.T) {
	room := &Room{
		ID:         "K7Q2",
		Name:       "Sala de Ana",
		MaxPlayers: 4,
		Turn:       2,
		Players: map[string]*players.Player{
			"p1": {
				Id:          "p1",
				Name:        "Ana",
				MoveTime:    1,
				Cards:       []structs.Card{},
				Investments: []int{},
				Alive:       true,
			},
		},
		Deck:          []structs.CardName{structs.CardAssassin},
		CurrentPlayer: "p1",
	}

	public := room.GetPublicRoomForUpdates()
	if public.Name != room.Name {
		t.Fatalf("name = %q, quer %q", public.Name, room.Name)
	}
	if public.MaxPlayers != 4 {
		t.Fatalf("maxPlayers = %d, quer 4", public.MaxPlayers)
	}
	if got := public.Players["p1"].MoveTime; got != 1 {
		t.Fatalf("moveTime do jogador = %d, quer 1", got)
	}

	raw, err := json.Marshal(public)
	if err != nil {
		t.Fatalf("erro ao serializar sala pública: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("erro ao reler sala pública: %v", err)
	}
	if payload["name"] != "Sala de Ana" {
		t.Fatalf("json name = %v, quer %q", payload["name"], "Sala de Ana")
	}
	if payload["maxPlayers"] != float64(4) {
		t.Fatalf("json maxPlayers = %v, quer 4", payload["maxPlayers"])
	}
}
