package messages

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocols"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocolsValidation"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/redis/go-redis/v9"
)

func ProcessPlay(ctx context.Context, rdb *redis.Client, roomID string, playerPlay *roomStructs.PlayerPlay, registryRules *rules.Registry) error {
	ok, err := room.AcquireRoomLock(ctx, rdb, roomID, config.InstanceID, 2*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	defer room.ReleaseRoomLock(ctx, rdb, roomID, config.InstanceID)

	roomData, err := room.LoadRoom(ctx, rdb, roomID)
	if err != nil {
		return err
	}

	switch playerPlay.Type {
	case roomStructs.PlayAssassinCard:
		assassinPayload, ok := playerPlay.Payload.(roomStructs.AssassinPayload)
		if !ok {
			return fmt.Errorf("payload does not match AssassinPayload structure")
		}

		if protocolsValidation.ValidateAssassinProtocol(roomData, playerPlay, assassinPayload, registryRules) {
			protocols.AssassinProtocol(ctx, rdb, roomData, playerPlay, assassinPayload)
		}
	}

	return nil
}
