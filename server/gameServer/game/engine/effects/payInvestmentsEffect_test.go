package effects

import (
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func TestCountdownInvestmentReturnsCoinWhenItExpires(t *testing.T) {
	room := &rooms.Room{}
	player := &players.Player{Coins: 0, Investments: []int{1, 3}}

	countdownInvestment(room, player, 1, 20)

	if player.Coins != 1 {
		t.Fatalf("moedas devolvidas = %d, quer 1", player.Coins)
	}
	if len(player.Investments) != 1 || player.Investments[0] != 2 {
		t.Fatalf("investimentos restantes = %v, quer [2]", player.Investments)
	}
}

func TestPayPlayerInvestmentsPaysHalfRoundedDown(t *testing.T) {
	room := &rooms.Room{}
	player := &players.Player{Coins: 0, Investments: []int{6, 6, 6, 6}}

	payPlayerInvestments(room, player, 20)

	if player.Coins != 2 {
		t.Fatalf("rendimento passivo = %d, quer 2 (metade de 4)", player.Coins)
	}
}
