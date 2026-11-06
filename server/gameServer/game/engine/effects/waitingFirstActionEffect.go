package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

// Efeito de espera pela primeira ação do jogador
func WaitingFirstActionEffect(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, currentPlayer *players.Player) error {

	if currentPlayer == nil {
		currentPlayer = &players.Player{Id: ""}
	}

	expiresAt := time.Now().Add(15 * time.Second).UTC()
	roomData.GameEvent = structs.NewGameEvent(currentPlayer.Id, structs.EventWaitingFirstAction, expiresAt, nil)

	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		return err
	}

	if err := roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{}); err != nil {
		return err
	}

	roomData.RegistryTimeout(rdb, ctx, expiresAt)
	return nil
}
