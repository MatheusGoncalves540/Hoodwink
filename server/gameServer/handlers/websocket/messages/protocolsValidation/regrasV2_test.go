package protocolsValidation

import (
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func loadTestRegistry(t *testing.T) *rules.Registry {
	t.Helper()

	gameRules, err := rules.LoadRules("../../../../game/rules/modes/ClassicRules.yaml")
	if err != nil {
		t.Fatalf("erro ao carregar ClassicRules.yaml: %v", err)
	}
	registry := rules.NewRegistry()
	registry.Register(string(structs.ClassicRules), gameRules)
	return registry
}

func newValidationRoom(player *players.Player, eventType structs.GameEventType, payload any) *rooms.Room {
	room := &rooms.Room{
		ID:                "TEST",
		Rules:             structs.ClassicRules,
		Tax:               0,
		CurrentPlayer:     player.Id,
		GameEvent:         structs.NewGameEvent(player.Id, eventType, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), payload),
		Players:           map[string]*players.Player{player.Id: player},
		DoubledCardValues: map[structs.TypePlayerPlays]structs.DoubledCardValues{},
	}
	return room
}

func newValidationPlayer() *players.Player {
	return &players.Player{
		Id:    "p1",
		Name:  "Ana",
		Coins: 10,
		Alive: true,
		Cards: []structs.Card{
			{Name: structs.CardAssassin, Index: 0},
			{Name: structs.CardGuardian, Index: 1, Dead: true},
		},
	}
}

func TestValidateCroupierAcceptsAliveAndDeadCards(t *testing.T) {
	registry := loadTestRegistry(t)

	for _, index := range []int{0, 1} {
		player := newValidationPlayer()
		room := newValidationRoom(player, structs.EventWaitingFirstAction, nil)

		idx := index
		play := &structs.PlayerPlay{
			Type:     structs.PlayCroupierCard,
			PlayerId: player.Id,
			Payload:  structs.NewCroupierPayload(&player.Id, &idx, nil),
		}

		if !ValidateCroupierProtocol(room, registry, play) {
			t.Fatalf("Croupier recusou a carta de índice %d (viva ou morta deveria ser aceita)", index)
		}
	}
}

func TestValidateCroupierRejectsInvalidPlays(t *testing.T) {
	registry := loadTestRegistry(t)

	player := newValidationPlayer()
	room := newValidationRoom(player, structs.EventWaitingFirstAction, nil)

	idx := 0
	twoCards := true
	otherPlayer := "p2"

	invalid := []struct {
		name string
		play *structs.PlayerPlay
	}{
		{
			name: "trocar duas cartas",
			play: &structs.PlayerPlay{
				Type:     structs.PlayCroupierCard,
				PlayerId: player.Id,
				Payload:  structs.NewCroupierPayload(&player.Id, &idx, &twoCards),
			},
		},
		{
			name: "carta de outro jogador",
			play: &structs.PlayerPlay{
				Type:     structs.PlayCroupierCard,
				PlayerId: player.Id,
				Payload:  structs.NewCroupierPayload(&otherPlayer, &idx, nil),
			},
		},
		{
			name: "índice inexistente",
			play: func() *structs.PlayerPlay {
				bad := 5
				return &structs.PlayerPlay{
					Type:     structs.PlayCroupierCard,
					PlayerId: player.Id,
					Payload:  structs.NewCroupierPayload(&player.Id, &bad, nil),
				}
			}(),
		},
	}

	for _, tc := range invalid {
		if ValidateCroupierProtocol(room, registry, tc.play) {
			t.Fatalf("Croupier aceitou jogada inválida: %s", tc.name)
		}
	}
}

func TestValidateGravediggerRequiresOwnDeadCard(t *testing.T) {
	registry := loadTestRegistry(t)

	player := newValidationPlayer()
	room := newValidationRoom(player, structs.EventWaitingFirstAction, nil)

	deadIndex := 1
	aliveIndex := 0
	otherPlayer := "p2"

	valid := &structs.PlayerPlay{
		Type:     structs.PlayGravediggerCard,
		PlayerId: player.Id,
		Payload:  structs.NewGravediggerPayload(&player.Id, &deadIndex),
	}
	if !ValidateGravediggerProtocol(room, registry, valid) {
		t.Fatal("Coveiro deveria aceitar uma carta morta do próprio jogador")
	}

	invalid := []struct {
		name    string
		payload structs.GravediggerPayload
	}{
		{"carta viva", structs.NewGravediggerPayload(&player.Id, &aliveIndex)},
		{"carta de outro jogador", structs.NewGravediggerPayload(&otherPlayer, &deadIndex)},
	}

	for _, tc := range invalid {
		play := &structs.PlayerPlay{
			Type:     structs.PlayGravediggerCard,
			PlayerId: player.Id,
			Payload:  tc.payload,
		}
		if ValidateGravediggerProtocol(room, registry, play) {
			t.Fatalf("Coveiro aceitou jogada inválida: %s", tc.name)
		}
	}
}

func TestValidateSelfishUsesTaxAdjustedThreshold(t *testing.T) {
	registry := loadTestRegistry(t)

	cases := []struct {
		name        string
		tax         int
		earnedCoins int
		want        bool
	}{
		{"tax 0, recebe 2", 0, 2, true},
		{"tax 0, recebe 1", 0, 1, false},
		{"tax +1, recebe 2", 1, 2, false},
		{"tax +1, recebe 3", 1, 3, true},
		{"tax -1, recebe 1", -1, 1, true},
	}

	for _, tc := range cases {
		player := newValidationPlayer()
		room := newValidationRoom(player, structs.EventEarnCoins, map[string]interface{}{
			"cause":       string(structs.EffectTrillionaire),
			"earnedCoins": float64(tc.earnedCoins),
		})
		room.Tax = tc.tax
		room.CurrentPlayer = "p2"
		room.GameEvent.PlayerID = "p2"

		play := &structs.PlayerPlay{
			Type:     structs.PlaySelfishCard,
			PlayerId: player.Id,
		}

		if got := ValidateSelfishProtocol(room, registry, play); got != tc.want {
			t.Fatalf("%s: Selfish aceitou=%v, quer %v", tc.name, got, tc.want)
		}
	}
}
