package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

// PassProtocol anota que o jogador não vai reagir à janela atual. Quando todos os
// elegíveis já responderam, a janela fecha na hora: o vencimento vai para agora e o
// processador resolve o próximo passo sem esperar o relógio (mesmo truque do
// lockAction).
func PassProtocol(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	if !roomData.RegisterPass(playerPlay.PlayerId) {
		return nil
	}

	now := roomData.Now().UTC()
	roomData.GameEvent.ExpiresAt = now
	sink.RegistryTimeout(ctx, roomData, now)
	return nil
}
