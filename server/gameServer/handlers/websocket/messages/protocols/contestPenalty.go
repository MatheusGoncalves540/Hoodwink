package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func ContestPenaltyProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, payload structs.ContestPenaltyPayload, chooserID string) error {
	// Remove o efeito de escolha aleatória criado na contestação.
	roomData.PopLastPendingEffect()

	// Enfileira o efeito de penalidade com a escolha feita pelo jogador.
	// O resolvedor aplica a morte e, quando aplicável, a troca da carta revelada.
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectContestPenalty, chooserID, payload)); err != nil {
		return err
	}

	// encerra a janela de input para evitar múltiplos cliques duplicados
	roomData.GameEvent = nil

	// agenda processamento imediato para consumir a fila
	sink.RegistryTimeout(ctx, roomData, roomData.Now().UTC())

	return nil
}
