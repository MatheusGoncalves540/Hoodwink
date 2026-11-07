package rooms

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/players"
)

// GetPlayer retorna o ponteiro do jogador pela playerId e um bool indicando se existe
func (r *Room) GetPlayer(playerId string) (*players.Player, error) {
	player, exists := r.Players[playerId]
	if !exists {
		return nil, fmt.Errorf("jogador alvo não encontrado: %s", playerId)
	}
	return player, nil
}

// IsFull indica se a sala já atingiu o número máximo de jogadores configurado.
func (r *Room) IsFull() bool {
	return r.MaxPlayers > 0 && len(r.Players) >= r.MaxPlayers
}

// GetNextRoundPlayer retorna o próximo jogador a jogar na rodada, considerando apenas jogadores vivos e conectados à sala.
// Retorna o próximo jogador, um booleano indicando se é o último jogador da rodada e um erro caso haja falha.
func (r *Room) GetNextRoundPlayer(currentPlayer *players.Player) (*players.Player, bool, bool, error) {
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

	for _, playerId := range r.SortedPlayerIDs() {
		player := r.Players[playerId]
		if !player.Alive || !r.IsConnected(playerId) {
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
func (r *Room) GetFirstAliveAndConnectedPlayer() *players.Player {
	var firstPlayer *players.Player

	for _, playerId := range r.SortedPlayerIDs() {
		player := r.Players[playerId]
		if !player.Alive || !r.IsConnected(playerId) {
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
func (r *Room) SelectRandomAliveAndConnectedPlayer() *players.Player {
	var selected *players.Player
	count := 0

	for _, playerId := range r.SortedPlayerIDs() {
		player := r.Players[playerId]
		if !player.Alive || !r.IsConnected(playerId) {
			continue
		}

		count++
		if r.Rand().IntN(count) == 0 {
			selected = player
		}
	}

	return selected
}

// VerifyIfPlayerDead verifica se um jogador está morto e atualiza seu status na sala, caso esteja morto
func (r *Room) VerifyIfPlayerDead(player *players.Player) {
	for _, card := range player.Cards {
		// caso todas as cartas do jogador estejam mortas, o jogador é considerado morto
		if card.Dead == false {
			return
		}
	}
	player.Alive = false
	player.ClearInvestments()

	r.UpdateGameOver()
}

// UpdateGameOver recalcula o estado de fim de jogo. O vencedor é o último jogador
// com cartas vivas; se ninguém restar, a partida termina sem vencedor.
func (r *Room) UpdateGameOver() {
	// Não define fim de jogo antes da partida começar
	if r.StartTime.IsZero() {
		return
	}

	aliveCount := 0
	winnerID := ""
	for _, playerId := range r.SortedPlayerIDs() {
		player := r.Players[playerId]
		if player.Alive {
			aliveCount++
			winnerID = player.Id
		}
	}

	r.GameOver = aliveCount <= 1
	if r.GameOver && aliveCount == 1 {
		r.WinnerID = winnerID
	} else {
		r.WinnerID = ""
	}
}
