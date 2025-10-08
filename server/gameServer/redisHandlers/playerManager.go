package redisHandlers

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// AddOrUpdatePlayerInRoom adiciona um jogador à sala ou atualiza seu status de conexão
func AddOrUpdatePlayerInRoom(ctx context.Context, rdb *redis.Client, roomId string, playerId string, username string) error {
	// Carrega a sala
	room, err := LoadRoom(ctx, rdb, roomId)
	if err != nil {
		return fmt.Errorf("erro ao carregar sala: %w", err)
	}

	// Procura se o jogador já existe na sala
	playerExists := false
	for i, player := range room.Players {
		if player.Id == playerId {
			// Jogador já existe, apenas atualiza status de conexão
			room.Players[i].Connected = true
			playerExists = true
			break
		}
	}

	// Se o jogador não existe, adiciona à sala
	if !playerExists {
		newPlayer := roomStructs.Player{
			Id:        playerId,
			Name:      username,
			Cards:     []string{},
			Coins:     2, // Valor inicial padrão do jogo
			Connected: true,
			Alive:     true,
		}
		room.Players = append(room.Players, newPlayer)
	}

	// Salva a sala atualizada
	return SaveRoom(ctx, rdb, room)
}

// RemovePlayerFromRoom remove um jogador da sala pelo seu playerId
func RemovePlayerFromRoom(ctx context.Context, rdb *redis.Client, roomId string, playerId string) error {
	// Carrega a sala
	room, err := LoadRoom(ctx, rdb, roomId)
	if err != nil {
		return fmt.Errorf("erro ao carregar sala: %w", err)
	}

	// Procura e remove o jogador
	playerFound := false
	newPlayers := make([]roomStructs.Player, 0, len(room.Players))
	for _, player := range room.Players {
		if player.Id == playerId {
			playerFound = true
			continue // Não adiciona o jogador removido
		}
		newPlayers = append(newPlayers, player)
	}

	if !playerFound {
		return fmt.Errorf("jogador %s não encontrado na sala %s", playerId, roomId)
	}

	room.Players = newPlayers

	// Salva a sala atualizada
	return SaveRoom(ctx, rdb, room)
}

// SetPlayerConnectionStatus atualiza apenas o status de conexão de um jogador
func SetPlayerConnectionStatus(ctx context.Context, rdb *redis.Client, roomId string, playerId string, connected bool) error {
	// Carrega a sala
	room, err := LoadRoom(ctx, rdb, roomId)
	if err != nil {
		return fmt.Errorf("erro ao carregar sala: %w", err)
	}

	// Procura e atualiza o jogador
	playerFound := false
	for i, player := range room.Players {
		if player.Id == playerId {
			room.Players[i].Connected = connected
			playerFound = true
			break
		}
	}

	if !playerFound {
		return fmt.Errorf("jogador %s não encontrado na sala %s", playerId, roomId)
	}

	// Salva a sala atualizada
	return SaveRoom(ctx, rdb, room)
}
