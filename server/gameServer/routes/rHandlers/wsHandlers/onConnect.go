package wsHandlers

import (
	"context"
	"log"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se conecta ao WebSocket
func OnConnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, roomId string, playerId string) {
	log.Println("Cliente conectado ao WebSocket")

	// Remove o TTL da sala quando um jogador se conecta
	err := redisHandlers.RemoveRoomTTL(ctx, rdb, roomId)
	if err != nil {
		utils.LogDebug("Erro ao remover TTL da sala: " + err.Error())
	} else {
		utils.LogDebug("TTL removido da sala " + roomId + " - sala agora é persistente")
	}

	// Adiciona ou atualiza o jogador na estrutura da sala
	err = redisHandlers.AddOrUpdatePlayerInRoom(ctx, rdb, roomId, playerId)
	if err != nil {
		utils.LogDebug("Erro ao adicionar/atualizar jogador na sala: " + err.Error())
	} else {
		utils.LogDebug("Jogador " + playerId + " adicionado/atualizado na sala " + roomId)
	}
}
