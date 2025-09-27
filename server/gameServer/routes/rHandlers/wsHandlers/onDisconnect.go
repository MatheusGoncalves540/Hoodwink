package wsHandlers

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se desconecta do WebSocket
func OnDisconnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, playerId string, roomId string) {
	redisHandlers.AcquirePlayerLock(ctx, rdb, playerId, 0)
	defer redisHandlers.ReleasePlayerLock(ctx, rdb, playerId)

	// Remove o registro do player na sala apenas se o jogo não estiver iniciado
	startTimeAny, err := redisHandlers.LoadRoomField(ctx, rdb, roomId, "StartTime")
	if err != nil {
		utils.LogDebug("Erro ao verificar horário de início da sala:" + err.Error())
		return
	}
	startTime, ok := startTimeAny.(time.Time)
	if !ok {
		utils.LogDebug("Erro: StartTime não é do tipo time.Time")
		return
	}
	if startTime.IsZero() {
		redisHandlers.UnregisterPlayerFromRoom(ctx, rdb, playerId)
	}
	utils.LogDebug("Cliente desconectado do WebSocket")
	// Aqui você pode adicionar lógica extra, como limpar recursos
}
