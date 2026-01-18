package protocols

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine/effects"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func ContestPenaltyProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, contestPenaltyPayload structs.ContestPenaltyPayload, sourcePlayerId string, targetPlayerId string) error {
	// extrai o player que está atacando
	sourcePlayer, err := roomData.GetPlayer(sourcePlayerId)
	if err != nil {
		return err
	}

	// extrai o player que está sendo atacado
	targetPlayer, err := roomData.GetPlayer(targetPlayerId)
	if err != nil {
		return err
	}

	// Remove o efeito de escolha randomica de carta
	roomData.CancelLastPendingEffect(ctx, rdb)

	// cria o efeito de penalidade por contestação
	effect := structs.Effect{
		Cause:        structs.EffectContestPenalty,
		SourcePlayer: sourcePlayer.Id,
		Payload: structs.KillCardPayload{
			Cause:           string(structs.EffectContestPenalty),
			TargetPlayer:    &targetPlayer.Id,
			TargetCardIndex: contestPenaltyPayload.TargetCardIndex,
		},
	}

	effects.KillCard(ctx, rdb, registryRules, roomData, effect)

	return nil
}
