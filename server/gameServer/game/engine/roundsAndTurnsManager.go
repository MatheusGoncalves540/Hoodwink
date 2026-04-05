package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func NextRound(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, registryRules *rules.Registry) error {
	currentPlayer, err := roomData.GetPlayer(roomData.CurrentPlayer)
	if err != nil {
		return err
	}

	// Define quem é o próximo jogador
	nextPlayer, isLast, isNewTurn, err := roomData.GetNextRoundPlayer(ctx, rdb, currentPlayer)
	if err != nil {
		return err
	}
	roomData.CurrentPlayer = nextPlayer.Id

	// limpa evento anterior
	roomData.GameEvent = nil

	// decrementa os rounds restantes para redução dos valores dobrados
	roomData.DecreaseDoubledCardValuesRounds()

	// Aleatoriamente aumenta ou diminui as taxas
	roomData.RandomlyChangeTaxes(ctx, rdb, registryRules)

	// executa o que for necessario para caso seja a ultima rodada do turno
	if isLast {
		LastRoundOfTurn(ctx, rdb, roomData, registryRules)
	}
	// inicia novo turno caso tenhamos voltado para o primeiro jogador do turno
	if isNewTurn {
		NewTurn(ctx, rdb, roomData, registryRules)
	}

	// inicia o efeito de espera pela primeira ação do próximo jogador
	if err := effects.WaitingFirstActionEffect(ctx, rdb, roomData, nextPlayer); err != nil {
		return err
	}

	// salva a sala com as atualizações de jogador atual e evento de espera pela primeira ação
	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		return err
	}

	return nil
}

// Função chamada no final de uma rodada
func LastRoundOfTurn(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, registryRules *rules.Registry) {
	effects.PayInvestments(ctx, rdb, registryRules, roomData)
}

// Função chamada no inicio de um novo turno
func NewTurn(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, registryRules *rules.Registry) {
	roomData.Turn++
}
