package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se conecta ao WebSocket
func OnConnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, roomId string, playerId string, username string) {
	utils.LogDebug("Cliente conectado ao WebSocket")

	// Remove o TTL da sala quando um jogador se conecta
	// err := room.RemoveRoomTTL(ctx, rdb, roomId)
	// if err != nil {
	// 	utils.LogDebug("Erro ao remover TTL da sala: " + err.Error())
	// } else {
	// 	utils.LogDebug("TTL removido da sala " + roomId + " - sala agora é persistente")
	// }

	// Adiciona ou atualiza o jogador na estrutura da sala
	err := room.AddPlayerInRoom(ctx, rdb, roomId, playerId, username)
	if err != nil {
		utils.LogError("Erro ao adicionar/atualizar jogador na sala: " + err.Error())
	} else {
		utils.LogDebug("Jogador " + playerId + " adicionado/atualizado na sala " + roomId)
	}

	roomData, err := room.LoadRoom(ctx, rdb, roomId)
	if err != nil {
		utils.LogError("Erro ao carregar dados da sala: " + err.Error())
		return
	}

	wsRoom.PublishRoomBroadcast(ctx, rdb, roomId, roomData)
}
