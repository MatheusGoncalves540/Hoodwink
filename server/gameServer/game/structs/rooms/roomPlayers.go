package rooms

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
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
		Id:       playerId,
		Name:     username,
		MoveTime: len(r.Players), // ordem de turnos baseada na ordem de entrada (primeiro a entrar tem moveTime 0, segundo tem moveTime 1, etc.)
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

// VerifyIfPlayerIsInEspecifiedRoom verifica se um jogador específico está registrado em uma sala específica (redis), retornando um booleano e um erro caso haja falha na consulta ao Redis
func (r *Room) VerifyIfPlayerIsInRegistredEspecifiedRoom(ctx context.Context, rdb *redis.Client, playerId string) (bool, error) {
	connectedToRoomID, registered, err := playerRedis.GetRegisteredRoomForPlayer(ctx, rdb, playerId)
	if err != nil {
		return false, fmt.Errorf("Falha ao obter sala registrada: %s", err.Error())
	}

	if registered && connectedToRoomID == r.ID {
		return true, nil
	}

	return false, nil
}

// GetNextRoundPlayer retorna o próximo jogador a jogar na rodada, considerando apenas jogadores vivos e conectados à sala.
// Retorna o próximo jogador, um booleano indicando se é o último jogador da rodada e um erro caso haja falha.
func (r *Room) GetNextRoundPlayer(ctx context.Context, rdb *redis.Client, currentPlayer *players.Player) (*players.Player, bool, bool, error) {
	if len(r.Players) == 0 {
		return nil, false, false, fmt.Errorf("nenhum jogador na sala %s", r.ID)
	}

	var nextPlayer *players.Player
	var minPlayer *players.Player
	var maxPlayer *players.Player

	foundNext := false
	foundMin := false
	foundMax := false

	currentMoveTime := currentPlayer.MoveTime

	for playerId, player := range r.Players {
		ok, err := r.VerifyIfPlayerIsInRegistredEspecifiedRoom(ctx, rdb, playerId)
		if !player.Alive || err != nil || !ok {
			continue
		}

		// menor MoveTime
		if !foundMin || player.MoveTime < minPlayer.MoveTime {
			minPlayer = player
			foundMin = true
		}

		// maior MoveTime
		if !foundMax || player.MoveTime > maxPlayer.MoveTime {
			maxPlayer = player
			foundMax = true
		}

		// próximo após o atual
		if player.MoveTime > currentMoveTime {
			if !foundNext || player.MoveTime < nextPlayer.MoveTime {
				nextPlayer = player
				foundNext = true
			}
		}
	}

	if !foundMin {
		return nil, false, false, fmt.Errorf("nenhum jogador válido na sala %s", r.ID)
	}

	// wrap-around
	if !foundNext {
		nextPlayer = minPlayer
	}

	// true quando o próximo for o ÚLTIMO
	isLast := nextPlayer.MoveTime == maxPlayer.MoveTime

	// troca de turno acontece quando voltamos para o primeiro
	isNewTurn := nextPlayer.MoveTime == minPlayer.MoveTime

	return nextPlayer, isLast, isNewTurn, nil
}

// GetFirstAliveAndConnectedPlayer retorna o primeiro jogador vivo e conectado encontrado na sala, ou nil caso não haja nenhum com base na ordem de MoveTime (ordem de entrada na sala)
func (r *Room) GetFirstAliveAndConnectedPlayer(ctx context.Context, rdb *redis.Client) *players.Player {
	var firstPlayer *players.Player

	for playerId, player := range r.Players {
		ok, err := r.VerifyIfPlayerIsInRegistredEspecifiedRoom(ctx, rdb, playerId)
		if !player.Alive || err != nil || !ok {
			continue
		}

		if firstPlayer == nil || player.MoveTime < firstPlayer.MoveTime {
			firstPlayer = player
		}
	}

	return firstPlayer
}

// SelectRandomAliveAndConnectedPlayer seleciona aleatoriamente um jogador vivo e conectado da sala
// retornando o jogador selecionado ou nil caso não haja nenhum jogador válido
func (r *Room) SelectRandomAliveAndConnectedPlayer(ctx context.Context, rdb *redis.Client) *players.Player {
	var selected *players.Player
	count := 0

	for playerId, player := range r.Players {
		ok, err := r.VerifyIfPlayerIsInRegistredEspecifiedRoom(ctx, rdb, playerId)
		if !player.Alive || err != nil || !ok {
			continue
		}

		count++
		if rand.IntN(count) == 0 {
			selected = player
		}
	}

	return selected
}
