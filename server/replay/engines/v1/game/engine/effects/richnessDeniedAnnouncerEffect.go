package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func RichnessDeniedAnnouncer(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	richnessDeniedPayload, ok := effect.Payload.(structs.RichnessDeniedPayload)
	if !ok {
		pure.LogError("payload inválido para RichnessDeniedPayload (announcer)")
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
