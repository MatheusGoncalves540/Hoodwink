package wsHandlers

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/wsMsgHandler"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada ao receber uma mensagem do cliente
func OnMessage(ctx context.Context, conn *websocket.Conn, rdb *redis.Client, evt *roomStructs.Event, claims jwt.MapClaims) {
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
	wsMsgHandler.ProcessWebSocketMessage(evt, ctx, rdb, room)
}
