package connection

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/player"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se desconecta do WebSocket
func OnDisconnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, playerId string, roomId string) {
	player.AcquirePlayerLock(ctx, rdb, playerId, 0)
	defer player.ReleasePlayerLock(ctx, rdb, playerId)

	// Remove o registro do player na sala apenas se o jogo não estiver iniciado
	startTimeAny, err := room.LoadRoomField(ctx, rdb, roomId, "StartTime")
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
		// Jogo não iniciado, remove a estrutura do player na sala
		err = room.RemovePlayerFromRoom(ctx, rdb, roomId, playerId)
		if err != nil {
			utils.LogDebug("Erro ao remover jogador da estrutura da sala: " + err.Error())
		}
	}

	// Verifica se a sala está vazia após a desconexão
	isEmpty, err := room.CheckIfRoomIsEmpty(ctx, rdb, roomId)
	if err != nil {
		utils.LogDebug("Erro ao verificar se sala está vazia: " + err.Error())
	} else if isEmpty {
		// Se a sala está vazia, define TTL de 20 segundos
		err = room.SetRoomTTL(ctx, rdb, roomId, 20*time.Second)
		if err != nil {
			utils.LogDebug("Erro ao definir TTL da sala vazia: " + err.Error())
		} else {
			utils.LogDebug("TTL de 20 segundos definido para sala vazia " + roomId)
		}
	}

	utils.LogDebug("Cliente desconectado do WebSocket")
}
