package room

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/handlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// Requisições chegando do websocket caem aqui.
//
// ProcessEvent gerencia o processamento de eventos em uma sala de jogo, garantindo exclusão mútua via lock no Redis.
//
// Dependendo do estado atual da sala (room.State), processa diferentes tipos de eventos:
// - rs.WaitingAction: processa ações dos jogadores, como uso de cartas.
// - rs.WaitingContest: processa contestações de jogadas.
// - rs.FinalizingAction: finaliza a ação atual e prepara para o próximo estado.
// - rs.TurnFinished: avança para o próximo turno e define o próximo jogador.
// Ao final, persiste o estado atualizado da sala no Redis.
// Retorna erro caso ocorra falha ao adquirir o lock, processar o evento ou salvar o estado.
func ProcessEvent(ctx context.Context, rdb *redis.Client, room *rs.Room, evt *eventQueue.Event) error {
	instanceID := utils.GetInstanceID()
	ok, err := redisHandlers.AcquireRoomLock(ctx, rdb, room.ID, instanceID, 5*time.Second)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("a sala %s está sendo modificada por outra instância, tente novamente", room.ID)
	}
	defer redisHandlers.ReleaseRoomLock(ctx, rdb, room.ID, instanceID)

	switch room.State {
	case rs.WaitingAction:
		if evt.Type == "action" {
			// Processa a ação do jogador
			var payloadMap map[string]interface{}
			payloadBytes, ok := evt.Payload.([]byte)
			if !ok {
				return fmt.Errorf("evt.Payload is not of type []byte")
			}
			if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
				return err
			}
			action, ok := payloadMap["action"].(string)
			if !ok {
				return nil
			}

			// Ação de uso de carta
			switch action {
			case "use_assassin":
				return handlers.UseAssassin(ctx, rdb, room, evt)
			default:
				// Adicionar mais ações conforme necessário
				return nil
			}
		}
	case rs.WaitingContest:
		if evt.Type == "contest" {
			// Processa a contestação
			payloadMap, ok := evt.Payload.(map[string]interface{})
			if !ok {
				return nil
			}
			contested, ok := payloadMap["contested"].(bool)
			if !ok {
				return nil
			}

			// Lida com a contestação da jogada
			return handlers.ProcessContest(ctx, rdb, room, evt, contested)
		}
	case rs.FinalizingAction:
		// Finaliza a ação
		room.State = rs.TurnFinished
	case rs.TurnFinished:
		// Avança para o próximo turno
		room.Turn++
		room.State = rs.WaitingAction
		// Adiciona lógica para definir o próximo jogador
	}
	return redisHandlers.SaveRoom(ctx, rdb, room)
}
