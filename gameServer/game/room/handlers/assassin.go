package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func UseAssassin(ctx context.Context, rdb *redis.Client, room *rs.Room, evt *eventQueue.Event) error {
	// Processa a ação de usar o Assassino
	var payload map[string]interface{}
	if raw, ok := evt.Payload.(json.RawMessage); ok {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("evt.Payload is not of type json.RawMessage")
	}
	target, ok := payload["target"].(string)
	if !ok {
		return fmt.Errorf("target não encontrado ou não é string")
	}

	// Cria o efeito pendente de matar uma carta
	effect := rs.Effect{
		Type:       "kill",
		From:       evt.PlayerId,
		To:         target,
		CardIndex:  -1,
		Executable: false,
		Reason:     "assassin_effect",
	}

	// Adiciona o efeito pendente
	room.CurrentMove = &rs.Move{
		PlayerId: evt.PlayerId,
		Action:   "use_assassin",
		TargetId: target,
	}
	room.PendingEffects = append(room.PendingEffects, effect)
	room.State = rs.WaitingContest

	// Agenda o próximo evento para o tempo de contestação (ex: 8 segundos)
	redisHandlers.ScheduleNextStep(ctx, rdb, room.ID, eventQueue.Event{
		Type:      "no_contest",
		PlayerId:  "system",
		TimeoutMs: 8000,
	})

	// Salva a sala
	return redisHandlers.SaveRoom(ctx, rdb, room)
}
