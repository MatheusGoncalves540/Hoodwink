package handlers

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

func UseAssassin(ctx context.Context, rdb *redis.Client, room *rs.Room, evt *eventQueue.Event) error {
	// Processa a ação de usar o Assassino
	payload := evt.Payload.(map[string]interface{})
	target := payload["target"].(string)

	// Cria o efeito pendente de matar uma carta
	effect := rs.Effect{
		Type:       "kill",
		From:       evt.PlayerUUID,
		To:         target,
		CardIndex:  -1,
		Executable: false,
		Reason:     "assassin_effect",
	}

	// Adiciona o efeito pendente
	room.CurrentMove = &rs.Move{
		PlayerUUID: evt.PlayerUUID,
		Action:     "use_assassin",
		TargetUUID: target,
	}
	room.PendingEffects = append(room.PendingEffects, effect)
	room.State = rs.WaitingContest

	// Agenda o próximo evento para o tempo de contestação (ex: 8 segundos)
	redisHandlers.ScheduleNextStep(ctx, rdb, room.ID, eventQueue.Event{
		Type:          "no_contest",
		PlayerUUID:    "system",
		TimeoutMillis: 8000,
	})

	// Salva a sala
	return redisHandlers.SaveRoom(ctx, rdb, room)
}
