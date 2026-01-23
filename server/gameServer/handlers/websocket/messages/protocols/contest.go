package protocols

import (
	"context"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/redis/go-redis/v9"
)

func ContestProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	// extrai o player do payload do evento atual (quem jogou a carta e está sendo contestado)
	ContestedPlayerId := roomData.GameEvent.PlayerID

	// extrai a carta do Evento atual (carta que está sendo contestada)
	ContestedCard := strings.TrimPrefix(string(roomData.GameEvent.Type), "CARD_PLAYED_")

	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "DisplayMessage")
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	contestPayload := structs.NewContestPayload(ContestedPlayerId, ContestedCard)

	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventContest, expiresAt, contestPayload)

	roomData.PendingEffects = append(roomData.PendingEffects,
		structs.Effect{
			Cause:        structs.EffectContest,
			SourcePlayer: sourcePlayer.Id,
			Payload:      contestPayload,
		},
	)

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
	return nil
}
