package engine

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/engine/effects/cardEffects"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

// resolveNextEffect é responsável por resolver o próximo efeito lógico pendente na sala. Ele é chamado após a resolução de um evento de apresentação ou efeito lógico, garantindo que os efeitos sejam processados em sequência.
func resolveNextEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room) {
	effectDto, ok := roomData.PopLastPendingEffect()
	if !ok {
		pure.LogError(fmt.Errorf("no pending effects to resolve"))
		return
	}

	effect, err := BuildEffect(effectDto)
	if err != nil {
		pure.LogError(err)
		return
	}

	switch effect.Cause {
	case structs.EffectCardKilled:
		effects.KillCard(ctx, sink, registryRules, roomData, effect)

	case structs.EffectEarnCoins:
		effects.EarnCoinsEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectStealCoins:
		effects.StealCoinsEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectContest:
		effects.ContestEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectContestPenalty:
		effects.ContestPenaltyEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectRevivedCard:
		effects.ReviveCardEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectChangedCard:
		effects.ChangeCardEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectInvestment:
		effects.InvestmentEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectRichnessDenied:
		effects.RichnessDeniedEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectAssassin:
		cardEffects.AssassinEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectKamikaze:
		cardEffects.KamikazeEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectTrillionaire:
		cardEffects.TrillionaireEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectPolitical:
		cardEffects.PoliticalEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectRebel:
		cardEffects.RebelEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectClairvoyant:
		cardEffects.ClairvoyantEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectGuardian:
		cardEffects.GuardianEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectTrickster:
		cardEffects.TricksterEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectGravedigger:
		cardEffects.GravediggerEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectInvestor:
		cardEffects.InvestorEffect(ctx, sink, registryRules, roomData, effect)

	case structs.EffectSelfish:
		cardEffects.SelfishEffect(ctx, sink, registryRules, roomData, effect)
	}

	sink.Broadcast(ctx, roomData, nil, []string{})
}
