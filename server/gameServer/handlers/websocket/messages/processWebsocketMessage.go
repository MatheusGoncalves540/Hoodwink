package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocols"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocolsValidation"
	"github.com/redis/go-redis/v9"
)

func ProcessPlay(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *roomStructs.PlayerPlay, registryRules *rules.Registry) error {
	ok, err := roomData.AcquireRoomLock(ctx, rdb, config.InstanceID, 2*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	defer roomData.ReleaseRoomLock(ctx, rdb, config.InstanceID)

	switch playerPlay.Type {
	case roomStructs.PlayAssassinCard:
		assassinPayload, ok := playerPlay.Payload.(roomStructs.AssassinPayload)
		if !ok {
			return fmt.Errorf("payload does not match AssassinPayload structure")
		}

		if protocolsValidation.ValidateAssassinProtocol(roomData, playerPlay, assassinPayload, registryRules) {
			protocols.AssassinProtocol(ctx, rdb, registryRules, roomData, playerPlay, assassinPayload)
		}
	}

	return nil
}
