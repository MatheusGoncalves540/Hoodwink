package room

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/handlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ProcessEvent(ctx context.Context, rdb *redis.Client, room *rs.Room, evt *eventQueue.Event) error {
	instanceID := utils.GetInstanceID()
	ok, err := redisHandlers.AcquireRoomLock(ctx, rdb, room.ID, instanceID, 5*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("não foi possível adquirir o lock da sala")
	}
	defer redisHandlers.ReleaseRoomLock(ctx, rdb, room.ID, instanceID)

	switch room.State {
	case rs.WaitingAction:
		if evt.Type == "action" {
			// Processa a ação do jogador
			payloadMap, ok := evt.Payload.(map[string]interface{})
			if !ok {
				return nil
			}
			action, ok := payloadMap["action"].(string)
			if !ok {
				return nil
			}

			// Ação de uso de carta
			switch action {
			case "use_assassin":
				return handlers.UseAssassin(ctx, rdb, room, evt)
			default:
				// Adicionar mais ações conforme necessário
				return nil
			}
		}
	case rs.WaitingContest:
		if evt.Type == "contest" {
			// Processa a contestação
			payloadMap, ok := evt.Payload.(map[string]interface{})
			if !ok {
				return nil
			}
			contested, ok := payloadMap["contested"].(bool)
			if !ok {
				return nil
			}

			// Lida com a contestação da jogada
			return handlers.ProcessContest(ctx, rdb, room, evt, contested)
		}
	case rs.FinalizingAction:
		// Finaliza a ação
		room.State = rs.TurnFinished
	case rs.TurnFinished:
		// Avança para o próximo turno
		room.Turn++
		room.State = rs.WaitingAction
		// Adiciona lógica para definir o próximo jogador
	}
	return redisHandlers.SaveRoom(ctx, rdb, room)
}
