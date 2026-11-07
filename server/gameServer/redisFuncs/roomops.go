package redisFuncs

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// AddPlayerInRoom adiciona um jogador à sala caso ele ainda não exista na estrutura da sala.
// Devolve aplicado (true quando o jogador foi adicionado). Não persiste: o WithRoom salva.
func AddPlayerInRoom(rulesRegistry *rules.Registry, room *rooms.Room, playerId string, username string) (bool, error) {
	player, err := room.GetPlayer(playerId)
	if err == nil && player != nil {
		return false, nil
	}

	if room.IsFull() {
		return false, fmt.Errorf("sala %s está cheia: %d jogadores (máximo %d)", room.ID, len(room.Players), room.MaxPlayers)
	}

	generalRules, err := room.GetGeneralRules(rulesRegistry)
	if err != nil {
		utils.LogError(err)
		return false, nil
	}

	room.Players[playerId] = &players.Player{
		Id:       playerId,
		Name:     username,
		MoveTime: room.NextMoveTime,
		Cards: []structs.Card{
			{Name: structs.CardKamikaze, Index: 0, Protected: false, Dead: false},
			{Name: structs.CardKamikaze, Index: 1, Protected: false, Dead: false},
		},
		Coins: *generalRules.InitialCoins,
		Alive: true,
	}
	room.NextMoveTime++
	room.AddConnected(playerId)

	return true, nil
}

// RemovePlayerFromRoom remove completamente um jogador da estrutura da sala.
// Devolve aplicado (true quando o jogador foi removido). Não persiste: o WithRoom salva.
func RemovePlayerFromRoom(room *rooms.Room, playerId string) (bool, error) {
	player, err := room.GetPlayer(playerId)
	if err != nil {
		return false, err
	}
	if player == nil {
		return false, fmt.Errorf("jogador %s não encontrado na sala %s", playerId, room.ID)
	}

	delete(room.Players, playerId)
	room.RemoveConnected(playerId)

	return true, nil
}
