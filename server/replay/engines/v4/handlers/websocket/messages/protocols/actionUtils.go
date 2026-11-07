package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/rooms"
)

// lockAction encerra a janela de primeira ação do turno e agenda o processamento
// imediato dos efeitos da ação básica, impedindo que o jogador realize outra ação
// antes do motor concluir a atual.
func lockAction(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, playerID string) {
	now := roomData.Now().UTC()
	roomData.GameEvent = structs.NewGameEvent(playerID, structs.EventActionTaken, now, nil)
	sink.RegistryTimeout(ctx, roomData, now)
}
