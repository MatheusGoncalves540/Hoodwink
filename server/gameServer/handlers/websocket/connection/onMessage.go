package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada ao receber uma mensagem do cliente
func OnMessage(ctx context.Context, conn *websocket.Conn, rdb *redis.Client, registryRules *rules.Registry, playerPlay *roomStructs.PlayerPlay, roomId string) {

	// Process WebSocket debug command - adm command
	messages.ProcessDebugCommand(ctx, rdb, roomId, playerPlay)
	// Process WebSocket message - player play
	messages.ProcessPlay(ctx, rdb, roomId, playerPlay, registryRules)
	// TODO Process WebSocket chat message
	// messages.ProcessChatMessage(playerPlay, ctx, rdb, roomId)
}
