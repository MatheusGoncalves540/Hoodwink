package messages

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

// ProcessPlay é o adapter fino do handler: delega a jogada ao motor puro
// (engine.ApplyPlay + broadcast via Sink) e devolve se a jogada foi aplicada.
func ProcessPlay(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, playerPlay *structs.PlayerPlay, registryRules *rules.Registry) bool {
	return engine.ProcessPlay(ctx, sink, roomData, playerPlay, registryRules)
}
