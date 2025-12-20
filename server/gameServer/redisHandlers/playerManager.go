package redisHandlers

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// AddPlayerInRoom adiciona um jogador à sala caso ele ainda não exista na estrutura da sala
func AddPlayerInRoom(ctx context.Context, rdb *redis.Client, roomId string, playerId string, username string) error {
	// Carrega a sala
	room, err := LoadRoom(ctx, rdb, roomId)
	if err != nil {
		return fmt.Errorf("erro ao carregar sala: %w", err)
	}

	// Adiciona o jogador à sala apenas se ainda não existir
	if _, exists := room.Players[playerId]; !exists {
		room.Players[playerId] = roomStructs.Player{
			Id:    playerId,
			Name:  username,
			Cards: []string{},
			Coins: 2, // Valor inicial padrão do jogo
			Alive: true,
		}
	}

	// Salva a sala atualizada
	return SaveRoom(ctx, rdb, room)
}

// RemovePlayerFromRoom remove completamente um jogador da estrutura da sala
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

	// Remove o jogador do mapa
	delete(room.Players, playerId)

	// Salva a sala atualizada
	return SaveRoom(ctx, rdb, room)
}
