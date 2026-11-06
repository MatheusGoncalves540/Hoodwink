package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func StealCoinsAnnouncer(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	stealCoinsPayload, ok := effect.Payload.(structs.StealCoinsPayload)
	if !ok {
		utils.LogError("payload inválido para StealCoinsPayload (announcer)")
		return nil
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		return err
	}

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventStealCoins, stealCoinsPayload)); err != nil {
		return err
	}

	// registra o efeito de perda de moedas
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectStealCoins, sourcePlayer.Id, stealCoinsPayload)); err != nil {
		return err
	}

	return nil
}
