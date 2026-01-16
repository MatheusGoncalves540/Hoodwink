package connection

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se desconecta do WebSocket
func OnDisconnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, roomId string, playerId string) {
	playerRedis.AcquirePlayerLock(ctx, rdb, playerId, 0)
	defer playerRedis.ReleasePlayerLock(ctx, rdb, playerId)

	roomData, err := redisFuncs.LoadRoom(ctx, rdb, roomId)
	if err != nil {
		utils.LogError("Erro ao carregar dados da sala no connect: " + err.Error())
		return
	}

	// Verifica se o jogo já foi iniciado
	if roomData.StartTime.IsZero() {
		// Jogo não iniciado, remove a estrutura do player na sala
		err := roomData.RemovePlayerFromRoom(ctx, rdb, playerId)
		if err != nil {
			utils.LogError("Erro ao remover jogador da estrutura da sala: " + err.Error())
		}
	}

	// Verifica se a sala está vazia após a desconexão
	isEmpty, err := roomData.CheckIfItsEmpty(ctx, rdb)
	if err != nil {
		utils.LogError("Erro ao verificar se sala está vazia: " + err.Error())
	} else if isEmpty {
		// Se a sala está vazia, define TTL de 20 segundos
		err = roomData.SetTTL(ctx, rdb, 20*time.Second)
		if err != nil {
			utils.LogError("Erro ao definir TTL da sala vazia: " + err.Error())
		} else {
			utils.LogDebug("TTL de 20 segundos definido para sala vazia " + roomData.ID)
		}
	}

	utils.LogDebug("Cliente desconectado do WebSocket")
	roomData.SendUpdatedRoomData(ctx, rdb)
}
