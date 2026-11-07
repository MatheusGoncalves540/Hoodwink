package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func NextRound(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, registryRules *rules.Registry) error {
	if roomData.GameOver {
		return nil
	}

	currentPlayer, err := roomData.GetPlayer(roomData.CurrentPlayer)
	if err != nil {
		return err
	}

	// Define quem é o próximo jogador
	nextPlayer, _, isNewTurn, err := roomData.GetNextRoundPlayer(currentPlayer)
	if err != nil {
		return err
	}
	roomData.CurrentPlayer = nextPlayer.Id

	// limpa evento anterior
	roomData.GameEvent = nil

	// decrementa os rounds restantes para redução dos valores dobrados
	roomData.DecreaseDoubledCardValuesRounds(registryRules)

	// Aleatoriamente aumenta ou diminui as taxas
	roomData.RandomlyChangeTaxes(registryRules)

	// O Investidor recebe no início do próprio turno e os investimentos avançam um turno
	effects.PayInvestmentsForPlayer(registryRules, roomData, nextPlayer)

	// o pagamento de investimentos pode matar o jogador ou encerrar a partida
	if roomData.GameOver || !nextPlayer.Alive {
		sink.Broadcast(ctx, roomData, nil, []string{})
		return nil
	}

	// inicia novo turno caso tenhamos voltado para o primeiro jogador do turno
	if isNewTurn {
		NewTurn(ctx, sink, roomData, registryRules)
	}

	// inicia o efeito de espera pela primeira ação do próximo jogador
	if err := effects.WaitingFirstActionEffect(ctx, sink, roomData, nextPlayer); err != nil {
		return err
	}

	return nil
}

// Função chamada no inicio de um novo turno
func NewTurn(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, registryRules *rules.Registry) {
	roomData.Turn++
}
