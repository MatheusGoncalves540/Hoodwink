package wsHandlers

import (
	"context"
	"log"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se desconecta do WebSocket
func OnDisconnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, playerId string, roomId string) {
	redisHandlers.AcquirePlayerLock(ctx, rdb, playerId, 0)
	defer redisHandlers.ReleasePlayerLock(ctx, rdb, playerId)

	redisHandlers.UnregisterPlayerFromRoom(ctx, rdb, playerId)
	log.Println("Cliente desconectado do WebSocket")
	// Aqui você pode adicionar lógica extra, como limpar recursos
}
