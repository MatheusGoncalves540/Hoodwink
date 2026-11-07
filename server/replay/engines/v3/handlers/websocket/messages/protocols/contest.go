package protocols

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

func ContestProtocol(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *structs.PlayerPlay) error {
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
	expiresAt := roomData.Now().Add(timeoutDuration * time.Second).UTC()

	contestPayload := structs.NewContestPayload(ContestedPlayerId, ContestedCard)

	roomData.GameEvent = structs.NewGameEvent(sourcePlayer.Id, structs.EventContest, expiresAt, contestPayload)

	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectContest, sourcePlayer.Id, contestPayload)); err != nil {
		return fmt.Errorf("falha ao registrar efeito de contestação: %w", err)
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
	return nil
}
