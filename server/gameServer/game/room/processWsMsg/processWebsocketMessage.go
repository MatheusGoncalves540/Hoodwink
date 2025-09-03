package processWsMsg

import (
	"context"

	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ProcessWebSocketMessage(evt *rs.Event, ctx context.Context, rdb *redis.Client, room *rs.Room) {
	switch room.State {
	case rs.WaitingGameStart:
		if evt.Type == "start" {
			err := OnStartAction(evt, ctx, rdb, room)
			if err != nil {
				utils.LogDebug("Error on process start: " + err.Error())
			}
		}
	case rs.WaitingFirstAction:
		if evt.Type == "action" {
			err := OnTypeAction(evt, ctx, rdb, room)
			if err != nil {
				utils.LogDebug("Error on process first action: " + err.Error())
			}
		}
		// if evt.Type == "" {
		//
		// }
		// case rs.WaitingContest:
		// 	if evt.Type == "contest" {
		// 		payloadMap, ok := evt.Payload.(map[string]interface{})
		// 		if !ok {
		// 			utils.PrintDebug("Payload contest inválido")
		// 			return
		// 		}
		// 		contested, ok := payloadMap["contested"].(bool)
		// 		if !ok {
		// 			utils.PrintDebug("Campo 'contested' inválido")
		// 			return
		// 		}

		// 		if err := handlers.ProcessContest(ctx, rdb, room, evt, contested); err != nil {
		// 			utils.PrintDebug("Erro ao processar contestação: " + err.Error())
		// 		}
		// 	}
		// case rs.FinalizingAction:
		// 	room.State = rs.TurnFinished
		// case rs.TurnFinished:
		// 	room.Turn++
		// 	room.State = rs.WaitingAction
	}
}
