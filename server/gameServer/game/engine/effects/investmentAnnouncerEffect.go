package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func InvestmentAnnouncer(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	InvestmentPayload, ok := effect.Payload.(structs.InvestmentPayload)
	if !ok {
		pure.LogError("payload inválido para InvestmentPayload")
		return nil
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		return err
	}

	// cria o payload de ganho de investimento
	investmentPayload := structs.NewInvestmentPayload(string(effect.Cause), InvestmentPayload.GracePeriod)

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventInvestment, investmentPayload)); err != nil {
		return err
	}

	// registra o efeito de ganho de investimento
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectInvestment, sourcePlayer.Id, investmentPayload)); err != nil {
		return err
	}

	return nil
}
