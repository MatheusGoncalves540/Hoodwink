package effects

import (
	"context"
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

	// Seleciona um jogador aleatóriamente
	selectedPlayer := roomData.GetFirstAliveAndConnectedPlayer(ctx, rdb)

	nextPlayer, _, _, err := roomData.GetNextRoundPlayer(ctx, rdb, selectedPlayer)
	if err != nil {
		return err
	}

	// Define o próximo jogador como o jogador atual da sala
	roomData.CurrentPlayer = nextPlayer.Id

	// Define startTime da sala para o horário atual, indicando que o jogo começou
	roomData.StartTime = time.Now().UTC()

	// inicia o efeito de espera pela primeira ação do próximo jogador
	if err := WaitingFirstActionEffect(ctx, rdb, roomData, nextPlayer); err != nil {
		return err
	}

	return nil
}
