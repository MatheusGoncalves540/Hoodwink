package wsHandlers

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/wsMsgHandler"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada ao receber uma mensagem do cliente
func OnMessage(ctx context.Context, conn *websocket.Conn, rdb *redis.Client, playerPlay *roomStructs.PlayerPlay, roomId string) {

	// Process WebSocket debug command - adm command
	wsMsgHandler.ProcessDebugCommand(playerPlay, ctx, rdb, roomId)
	// Process WebSocket message - player play
	wsMsgHandler.ProcessPlay(playerPlay, ctx, rdb)
}
