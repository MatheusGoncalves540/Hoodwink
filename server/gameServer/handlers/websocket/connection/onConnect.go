package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se conecta ao WebSocket
func OnConnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, roomId string, playerId string, username string) {
	roomData, err := redisFuncs.LoadRoom(ctx, rdb, roomId)
	if err != nil {
		utils.LogError("Erro ao carregar dados da sala no connect: " + err.Error())
		return
	}

	err = roomData.AddPlayerInRoom(ctx, rdb, playerId, username)
	if err != nil {
		utils.LogError("Erro ao adicionar/atualizar jogador na sala: " + err.Error())
	}

	utils.LogDebug("Cliente conectado ao WebSocket")
	roomData.SendUpdatedRoomData(ctx, rdb)
}
