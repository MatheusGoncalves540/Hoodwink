package protocols

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

// lockAction encerra a janela de primeira ação do turno e agenda o processamento
// imediato dos efeitos da ação básica, impedindo que o jogador realize outra ação
// antes do motor concluir a atual.
func lockAction(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerID string) {
	now := time.Now().UTC()
	roomData.GameEvent = structs.NewGameEvent(playerID, structs.EventActionTaken, now, nil)
	roomData.RegistryTimeout(rdb, ctx, now)
}
