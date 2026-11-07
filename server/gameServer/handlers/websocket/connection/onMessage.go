package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/redis/go-redis/v9"
)

// OnMessage roteia a mensagem do cliente e devolve aplicado (true quando algo mudou).
// Não persiste: o WithRoom salva quando aplicado.
func OnMessage(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, playerPlay *structs.PlayerPlay, roomData *rooms.Room) bool {
	sink := redisFuncs.NewSink(rdb)

	aplicado := false

	// Start legítimo da partida (botão de abrir as cortinas).
	if ProcessStartCommand(ctx, rdb, roomData, playerPlay) {
		aplicado = true
	}

	// Process WebSocket debug command - adm command
	if ProcessDebugCommand(ctx, rdb, roomData, playerPlay) {
		aplicado = true
	}
	// Process WebSocket message - player play
	if messages.ProcessPlay(ctx, sink, roomData, playerPlay, registryRules) {
		aplicado = true
	}
	// Process WebSocket chat message
	if ProcessChatMessage(ctx, rdb, roomData, playerPlay) {
		aplicado = true
	}

	return aplicado
}
