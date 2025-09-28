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

	// Marca o jogador como desconectado na estrutura da sala
	err = redisHandlers.SetPlayerConnectionStatus(ctx, rdb, roomId, playerId, false)
	if err != nil {
		utils.LogDebug("Erro ao marcar jogador como desconectado: " + err.Error())
	}

	// Verifica se a sala está vazia após a desconexão
	isEmpty, err := redisHandlers.CheckIfRoomIsEmpty(ctx, rdb, roomId)
	if err != nil {
		utils.LogDebug("Erro ao verificar se sala está vazia: " + err.Error())
	} else if isEmpty {
		// Se a sala está vazia, define TTL de 20 segundos
		err = redisHandlers.SetRoomTTL(ctx, rdb, roomId, 20*time.Second)
		if err != nil {
			utils.LogDebug("Erro ao definir TTL da sala vazia: " + err.Error())
		} else {
			utils.LogDebug("TTL de 20 segundos definido para sala vazia " + roomId)
		}
	}

	utils.LogDebug("Cliente desconectado do WebSocket")
}
