package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada ao receber uma mensagem do cliente
func OnMessage(ctx context.Context, conn *websocket.Conn, rdb *redis.Client, playerPlay *roomStructs.PendingEvent, roomId string) {

	// Process WebSocket debug command - adm command
	messages.ProcessDebugCommand(playerPlay, ctx, rdb, roomId)
	// Process WebSocket message - player play
	messages.ProcessPlay(playerPlay, ctx, rdb)
}
