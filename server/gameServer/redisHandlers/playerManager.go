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
	if _, exists := room.Players[playerId]; exists {
		// Jogador já existe, apenas atualiza status de conexão
		player := room.Players[playerId]
		player.Connected = true
		room.Players[playerId] = player
	} else {
		// Se o jogador não existe, adiciona à sala
		newPlayer := roomStructs.Player{
			Id:        playerId,
			Name:      username,
			Cards:     []string{},
			Coins:     2, // Valor inicial padrão do jogo
			Connected: true,
			Alive:     true,
		}
		room.Players[playerId] = newPlayer
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

	// Verifica se o jogador existe
	if _, exists := room.Players[playerId]; !exists {
		return fmt.Errorf("jogador %s não encontrado na sala %s", playerId, roomId)
	}

	delete(room.Players, playerId)

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

	// Verifica se o jogador existe
	if _, exists := room.Players[playerId]; !exists {
		return fmt.Errorf("jogador %s não encontrado na sala %s", playerId, roomId)
	}

	player := room.Players[playerId]
	player.Connected = connected
	room.Players[playerId] = player

	// Salva a sala atualizada
	return SaveRoom(ctx, rdb, room)
}
