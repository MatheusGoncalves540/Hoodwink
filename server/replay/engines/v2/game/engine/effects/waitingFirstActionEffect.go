package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/rooms"
)

// Efeito de espera pela primeira ação do jogador
func WaitingFirstActionEffect(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, currentPlayer *players.Player) error {
	sink = rooms.SinkOrNoop(sink)

	if currentPlayer == nil {
		currentPlayer = &players.Player{Id: ""}
	}

	expiresAt := roomData.Now().Add(15 * time.Second).UTC()
	roomData.GameEvent = structs.NewGameEvent(currentPlayer.Id, structs.EventWaitingFirstAction, expiresAt, nil)

	if err := sink.Broadcast(ctx, roomData, nil, []string{}); err != nil {
		return err
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
	return nil
}
