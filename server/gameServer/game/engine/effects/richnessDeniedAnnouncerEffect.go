package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func RichnessDeniedAnnouncer(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	richnessDeniedPayload, ok := effect.Payload.(structs.RichnessDeniedPayload)
	if !ok {
		utils.LogError("payload inválido para RichnessDeniedPayload (announcer)")
		return nil
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		return err
	}

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventRichnessDenied, richnessDeniedPayload)); err != nil {
		return err
	}

	// registra o efeito de negação de riqueza
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectRichnessDenied, sourcePlayer.Id, richnessDeniedPayload)); err != nil {
		return err
	}

	return nil
}
