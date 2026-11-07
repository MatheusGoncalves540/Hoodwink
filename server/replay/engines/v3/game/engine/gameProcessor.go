package engine

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

// Tick fecha a janela atual e avança a sala um passo: próximo evento de apresentação,
// próximo efeito lógico ou próximo turno. Com sink no-op, roda só em memória (testes de cenário).
func Tick(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room) {
	sink = rooms.SinkOrNoop(sink)

	roomData.GameEvent = nil

	if roomData.HasPendingPresentationEvent() {
		createGameEventFromPresentationEvent(ctx, sink, registryRules, roomData)
		return
	}

	// Resolve próximo efeito, se existir
	if roomData.HasPendingLogicEffect() {
		resolveNextEffect(ctx, sink, registryRules, roomData)
		scheduleImmediateProcessing(ctx, sink, roomData)
		return
	}

	// Nada pendente → próximo turno (ou fim de jogo)
	if roomData.GameEvent == nil && !roomData.HasPendingLogicEffect() && !roomData.HasPendingPresentationEvent() {
		roomData.UpdateGameOver()
		if roomData.GameOver {
			sink.ClearTimeout(ctx, roomData)
			sink.Broadcast(ctx, roomData, nil, []string{})
			return
		}
		NextRound(ctx, sink, roomData, registryRules)
	}
}

func createGameEventFromPresentationEvent(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room) {
	presentationEvent, ok := roomData.PopNextPendingPresentationEvent()
	if !ok {
		return
	}

	timeoutField := presentationEvent.TimeoutField
	if timeoutField == "" {
		timeoutField = "DisplayMessage"
	}

	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, timeoutField)
	if err != nil {
		pure.LogError(err)
		return
	}

	expiresAt := roomData.Now().Add(timeoutDuration * time.Second).UTC()
	roomData.GameEvent = structs.NewGameEvent(presentationEvent.PlayerID, presentationEvent.Type, expiresAt, presentationEvent.Payload)

	var (
		confidencialRoomData *rooms.Room
		playersThatCanSee    []string
	)

	if presentationEvent.ConfidencialPayload != nil && len(presentationEvent.ConfidencialPlayerIds) > 0 {
		confidencialRoomData = roomData.Clone()
		confidencialRoomData.GameEvent.Payload = presentationEvent.ConfidencialPayload
		playersThatCanSee = append(playersThatCanSee, presentationEvent.ConfidencialPlayerIds...)
	}

	if err := sink.Broadcast(ctx, roomData, confidencialRoomData, playersThatCanSee); err != nil {
		pure.LogError(err)
		return
	}

	sink.RegistryTimeout(ctx, roomData, expiresAt)
}

// scheduleImmediateProcessing verifica se a sala tem um evento de apresentação ou efeito lógico pendente e registra um timeout imediato para processamento.
func scheduleImmediateProcessing(ctx context.Context, sink rooms.Sink, roomData *rooms.Room) {
	if roomData.GameEvent != nil {
		return
	}
	if !roomData.HasPendingPresentationEvent() && !roomData.HasPendingLogicEffect() {
		return
	}

	sink.RegistryTimeout(ctx, roomData, roomData.Now().UTC())
}
