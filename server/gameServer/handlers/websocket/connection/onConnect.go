package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se conecta ao WebSocket
func OnConnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerId string, username string) {
	err := roomData.AddPlayerInRoom(ctx, rdb, playerId, username)
	if err != nil {
		utils.LogError("Erro ao adicionar/atualizar jogador na sala: " + err.Error())
	}

	utils.LogDebug("Cliente conectado ao WebSocket")
	roomData.PublishRoomBroadcast(ctx, rdb, roomData)
}
