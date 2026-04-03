package rooms

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/redis/go-redis/v9"
)

// GetPlayer retorna o ponteiro do jogador pela playerId e um bool indicando se existe
func (r *Room) GetPlayer(playerId string) (*players.Player, error) {
	player, exists := r.Players[playerId]
	if !exists {
		return nil, fmt.Errorf("jogador alvo não encontrado: %s", playerId)
	}
	return player, nil
}

// AddPlayerInRoom adiciona um jogador à sala caso ele ainda não exista na estrutura da sala
func (r *Room) AddPlayerInRoom(ctx context.Context, rdb *redis.Client, playerId string, username string) error {
	// Verifica se o jogador já existe
	player, err := r.GetPlayer(playerId)
	if err == nil && player != nil {
		// Jogador já existe, nada a fazer
		return nil
	}

	// Adiciona o jogador ao mapa
	r.Players[playerId] = &players.Player{
		Id:   playerId,
		Name: username,
		Cards: []structs.Card{
			{Name: structs.CardKamikaze, Index: 0, Protected: false, Dead: false},
			{Name: structs.CardKamikaze, Index: 1, Protected: false, Dead: false},
		}, // TODO remover valor fixo de teste
		Coins: 10, // Valor inicial padrão do jogo
		Alive: true,
	}

	// Salva a sala atualizada
	err = r.SaveRoom(ctx, rdb)
	return nil
}

// RemovePlayerFromRoom remove completamente um jogador da estrutura da sala
func (r *Room) RemovePlayerFromRoom(ctx context.Context, rdb *redis.Client, playerId string) error {
	// Verifica se o jogador existe
	player, err := r.GetPlayer(playerId)
	if err != nil {
		return err
	}
	if player == nil {
		return fmt.Errorf("jogador %s não encontrado na sala %s", playerId, r.ID)
	}

	// Remove o jogador do mapa
	delete(r.Players, playerId)

	// Salva a sala atualizada
	return r.SaveRoom(ctx, rdb)
}
