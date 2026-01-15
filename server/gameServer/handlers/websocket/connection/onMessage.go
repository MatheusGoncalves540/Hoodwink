package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages"
	"github.com/redis/go-redis/v9"
)

// Função chamada ao receber uma mensagem do cliente
func OnMessage(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, playerPlay *roomStructs.PlayerPlay, roomData *rooms.Room) {

	// Process WebSocket debug command - adm command
	messages.ProcessDebugCommand(ctx, rdb, roomData, playerPlay)
	// Process WebSocket message - player play
	messages.ProcessPlay(ctx, rdb, roomData, playerPlay, registryRules)
	// TODO Process WebSocket chat message
	// messages.ProcessChatMessage(playerPlay, ctx, rdb, roomId)
}
