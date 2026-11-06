package effects

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

// Efeito de espera pela primeira ação do jogador
func StartGameEffect(ctx context.Context, rdb *redis.Client, roomData *rooms.Room) error {
	// troca as cartas de todos os jogadores para iniciar o jogo
	err := roomData.ChangeAllCards()
	if err != nil {
		return err
	}

	// Seleciona o primeiro jogador vivo e conectado na ordem da sala
	selectedPlayer := roomData.GetFirstAliveAndConnectedPlayer(ctx, rdb)
	if selectedPlayer == nil {
		return fmt.Errorf("nenhum jogador válido para iniciar a partida")
	}

	// Define o primeiro jogador como o jogador atual da sala
	roomData.CurrentPlayer = selectedPlayer.Id
	roomData.GameOver = false
	roomData.WinnerID = ""

	// Define startTime da sala para o horário atual, indicando que o jogo começou
	roomData.StartTime = time.Now().UTC()

	// inicia o efeito de espera pela primeira ação do próximo jogador
	if err := WaitingFirstActionEffect(ctx, rdb, roomData, selectedPlayer); err != nil {
		return err
	}

	return nil
}
