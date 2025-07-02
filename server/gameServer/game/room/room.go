package room

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers.go"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func ProcessEvent(ctx context.Context, rdb *redis.Client, room *rs.Room, evt *Event) error {
	switch room.State {
	case rs.WaitingAction:
		if evt.Type == "action" {
			payloadMap, ok := evt.Payload.(map[string]interface{})
			if !ok {
				return nil
			}
			action, ok := payloadMap["action"].(string)
			if !ok {
				return nil
			}
			room.CurrentMove = &rs.Move{
				PlayerUUID: evt.PlayerUUID,
				Action:     action,
			}
			room.State = rs.WaitingContest
		}
	case rs.WaitingContest:
		if evt.Type == "contest" {
			payloadMap, ok := evt.Payload.(map[string]interface{})
			if !ok {
				return nil
			}
			contested, ok := payloadMap["contested"].(bool)
			if !ok {
				return nil
			}
			if contested {
				room.State = rs.ResolvingContest
			} else {
				room.State = rs.FinalizingAction
			}
		}
	case rs.ResolvingContest:
		// lógica de resolução da contestação
	case rs.WaitingKamikazeResponse:
		// lógica de kamikaze
	case rs.FinalizingAction:
		room.State = rs.TurnFinished
	case rs.TurnFinished:
		room.Turn++
		room.State = rs.WaitingAction
		// lógica para definir próximo jogador
	}
	return redisHandlers.SaveRoom(ctx, rdb, room)
}
