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
	roomData.PopLastPendingEffect()

	// cria o efeito de penalidade por contestação
	killPayload := structs.NewKillCardPayload(string(structs.EffectContestPenalty), &targetPlayer.Id, contestPenaltyPayload.TargetCardIndex)

	killEffect := structs.Effect{
		Cause:        structs.EffectContestPenalty,
		SourcePlayer: sourcePlayer.Id,
		Payload:      killPayload,
	}

	effects.KillAnnouncerCard(ctx, rdb, registryRules, roomData, killEffect)

	return nil
}
