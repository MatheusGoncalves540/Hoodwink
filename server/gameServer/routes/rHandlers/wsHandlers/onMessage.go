package wsHandlers

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/processWsMsg"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada ao receber uma mensagem do cliente
func OnMessage(ctx context.Context, conn *websocket.Conn, rdb *redis.Client, evt *rs.Event, claims jwt.MapClaims) {
	// Verify if event is valid
	if evt == nil {
		utils.LogDebug("Evento inválido")
		return
	}

	room, err := redisHandlers.LoadRoom(ctx, rdb, evt.RoomId)
	if err != nil {
		utils.LogDebug("Erro ao buscar sala: " + err.Error())
		return
	}

	instanceID := utils.GetInstanceID()
	ok, err := redisHandlers.AcquireRoomLock(ctx, rdb, room.ID, instanceID, 5*time.Second)
	if err != nil {
		utils.LogDebug("Erro ao adquirir lock da sala: " + err.Error())
		return
	}
	if !ok {
		utils.LogDebug("a sala " + room.ID + " está sendo modificada por outra instância, tente novamente")
		return
	}
	defer redisHandlers.ReleaseRoomLock(ctx, rdb, room.ID, instanceID)

	// Process WebSocket message
	switch room.State {
	case rs.WaitingGameStart:
		if evt.Type == "start" {
			err := processWsMsg.OnStartAction(evt, ctx, rdb, room)
			if err != nil {
				utils.LogDebug("Error on process start: " + err.Error())
			}
		}
	case rs.WaitingFirstAction:
		if evt.Type == "action" {
			err := processWsMsg.OnTypeAction(evt, ctx, rdb, room)
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
