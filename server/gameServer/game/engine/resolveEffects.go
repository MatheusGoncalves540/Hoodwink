package engine

import (
	"context"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects/cardEffects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// resolveNextEffect é responsável por resolver o próximo efeito lógico pendente na sala. Ele é chamado após a resolução de um evento de apresentação ou efeito lógico, garantindo que os efeitos sejam processados em sequência.
func resolveNextEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room) {
	effectDto, ok := roomData.PopLastPendingEffect()
	if !ok {
		utils.LogError(fmt.Errorf("no pending effects to resolve"))
		return
	}

	effect, err := BuildEffect(effectDto)
	if err != nil {
		utils.LogError(err)
		return
	}

	switch effect.Cause {
	case structs.EffectCardKilled:
		effects.KillCard(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectEarnCoins:
		effects.EarnCoinsEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectStealCoins:
		effects.StealCoinsEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectContest:
		effects.ContestEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectContestPenalty:
		effects.ContestPenaltyEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectRevivedCard:
		effects.ReviveCardEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectChangedCard:
		effects.ChangeCardEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectInvestment:
		effects.InvestmentEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectRichnessDenied:
		effects.RichnessDeniedEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectAssassin:
		cardEffects.AssassinEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectKamikaze:
		cardEffects.KamikazeEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectTrillionaire:
		cardEffects.TrillionaireEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectPolitical:
		cardEffects.PoliticalEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectRebel:
		cardEffects.RebelEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectClairvoyant:
		cardEffects.ClairvoyantEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectGuardian:
		cardEffects.GuardianEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectTrickster:
		cardEffects.TricksterEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectGravedigger:
		cardEffects.GravediggerEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectInvestor:
		cardEffects.InvestorEffect(ctx, rdb, registryRules, roomData, effect)

	case structs.EffectSelfish:
		cardEffects.SelfishEffect(ctx, rdb, registryRules, roomData, effect)
	}

	roomData.SaveRoom(ctx, rdb)
	roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{})
}
