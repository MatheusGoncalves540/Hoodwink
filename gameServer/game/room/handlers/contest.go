package handlers

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/redis/go-redis/v9"
)

// Processa a contestação da jogada
func ProcessContest(ctx context.Context, rdb *redis.Client, room *rs.Room, evt *eventQueue.Event, contested bool) error {
	if contested {
		// A contestação foi válida. Verifica se o jogador realmente tinha a carta.
		hasCard := true // Aqui, você deve verificar se o jogador realmente tem a carta. Simplificado aqui.

		if hasCard {
			// Jogador tinha a carta
			// O efeito de matar a carta pode ser executado
			effect := rs.Effect{
				Type:       "kill",
				From:       evt.PlayerId,
				To:         room.CurrentMove.TargetId,
				CardIndex:  -1,
				Executable: true,
			}
			room.PendingEffects = append(room.PendingEffects, effect)
		}
	} else {
		// Contestação errada
		// O jogador que contestou perde uma carta
		// Você pode aplicar a lógica para remover a carta do jogador que fez a contestação errada
	}

	// Finaliza a ação e avança para o próximo turno
	room.State = rs.FinalizingAction
	return redisHandlers.SaveRoom(ctx, rdb, room)
}
