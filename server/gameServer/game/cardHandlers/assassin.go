package cardHandlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisHandlers"
	"github.com/redis/go-redis/v9"
)

func UseAssassin(ctx context.Context, rdb *redis.Client, room *roomStructs.Room, evt *roomStructs.Event) error {
	// Processa a ação de usar o Assassino
	var payload map[string]any
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
	effect := roomStructs.Effect{
		Type:       "kill",
		From:       evt.PlayerId,
		To:         target,
		CardIndex:  -1,
		Executable: false,
		Reason:     "assassin_effect",
	}

	// Adiciona o efeito pendente
	room.CurrentMove = &roomStructs.Move{
		PlayerId: evt.PlayerId,
		Action:   "use_assassin",
		TargetId: target,
	}
	room.PendingEffects = append(room.PendingEffects, effect)
	room.State = roomStructs.WaitingContest

	// Agenda o próximo evento para o tempo de contestação (ex: 8 segundos)
	redisHandlers.ScheduleNextStep(ctx, rdb, room.ID, roomStructs.Event{
		Type:      "no_contest",
		PlayerId:  "system",
		TimeoutMs: 8000,
	})

	// Salva a sala
	return redisHandlers.SaveRoom(ctx, rdb, room)
}
