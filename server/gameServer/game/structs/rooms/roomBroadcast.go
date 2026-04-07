package rooms

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// BroadcastMsgToRoom publica mensagem para todos os players de uma sala
func (r *Room) BroadcastMsgToRoom(ctx context.Context, rdb *redis.Client, message any) error {
	if pubMsg, err := json.Marshal(message); err == nil {
		return rdb.Publish(ctx, "room:"+r.ID+":broadcast", pubMsg).Err()
	}
	return nil
}

// SendUpdatedRoomData publica nova versão da sala para todos os players ou para players específicos.
// Se confidencialRoomData for nil OU confidencialPlayerIds estiver vazio, envia para todos (toAll: true).
// Caso contrário, envia apenas para os playerIds especificados usando confidencialRoomData.
func (r *Room) SendUpdatedRoomData(ctx context.Context, rdb *redis.Client, confidencialRoomData *Room, confidencialPlayerIds []string) error {
	// Verifica se é broadcast para todos ou seletivo
	toAll := confidencialRoomData == nil || len(confidencialPlayerIds) == 0

	if toAll {
		// Broadcast para todos - usa roomData pública normal
		if utils.IsUnsafeDebugMode() {
			return r.PublishUnsafeRoomData(ctx, rdb, true, nil)
		} else {
			return r.PublishRoomUpdate(ctx, rdb, true, nil)
		}
	} else {
		// Broadcast para todos - usa roomData pública normal
		// depois envia Broadcast seletivo - usa confidencialRoomData
		if utils.IsUnsafeDebugMode() {
			r.PublishUnsafeRoomData(ctx, rdb, true, nil)
			return confidencialRoomData.PublishUnsafeRoomData(ctx, rdb, false, confidencialPlayerIds)
		} else {
			r.PublishRoomUpdate(ctx, rdb, true, nil)
			return confidencialRoomData.PublishRoomUpdate(ctx, rdb, false, confidencialPlayerIds)
		}
	}
}

// PublishUnsafeRoomData publica a versão privada completa, não segura da sala para todos os players de uma sala (APENAS PARA DEBUG)
func (r *Room) PublishUnsafeRoomData(ctx context.Context, rdb *redis.Client, toAll bool, playerIds []string) error {
	var broadcastMsg *structs.BroadcastMessage

	if toAll {
		broadcastMsg = structs.NewBroadcastMessage(r)
	} else {
		broadcastMsg = structs.NewSelectiveBroadcastMessage(r, playerIds)
	}

	if pubMsg, err := broadcastMsg.ToJSON(); err == nil {
		return rdb.Publish(ctx, "room:"+r.ID+":broadcast", pubMsg).Err()
	}
	return nil
}

// PublishRoomUpdate publica a versão pública da sala para todos os players de uma sala
func (r *Room) PublishRoomUpdate(ctx context.Context, rdb *redis.Client, toAll bool, playerIds []string) error {
	// Prepara os dados públicos da sala
	roomDataPublic := PublicRoomForUpdates{
		ID:                r.ID,
		Turn:              r.Turn,
		Tax:               r.Tax,
		Players:           make(map[string]players.PublicPlayerForUpdates, len(r.Players)),
		Deck:              len(r.Deck),
		CurrentPlayer:     r.CurrentPlayer,
		GameEvent:         r.GameEvent,
		PendingEffects:    r.PendingEffects,
		GameOver:          r.GameOver,
		StartTime:         r.StartTime,
		DoubledCardValues: r.DoubledCardValues,
	}

	// Prepara os dados públicos dos jogadores
	for playerId, player := range r.Players {
		publicPlayer := player.GetPublicPlayerForUpdates()
		roomDataPublic.Players[playerId] = publicPlayer
	}

	// Cria a mensagem de broadcast
	var broadcastMsg *structs.BroadcastMessage
	if toAll {
		broadcastMsg = structs.NewBroadcastMessage(roomDataPublic)
	} else {
		broadcastMsg = structs.NewSelectiveBroadcastMessage(roomDataPublic, playerIds)
	}

	// Publica a mensagem
	if pubMsg, err := broadcastMsg.ToJSON(); err == nil {
		rdb.Publish(ctx, "room:"+r.ID+":broadcast", pubMsg).Err()
	}

	// Sempre que for enviar um broadcast, mesmo que seja para todos, envia também a versão privada para cada player individualmente
	// para garantir que cada player receba os dados confidenciais que só ele pode ver (como as cartas da mão)
	for _, player := range r.Players {
		roomDataPrivate := roomDataPublic.Clone()
		roomDataPrivate.Players[player.Id] = player.GetPrivatePlayerForPublicUpdates()

		privatePlayersInfo := structs.NewSelectiveBroadcastMessage(roomDataPrivate, []string{player.Id})

		// Publica a mensagem
		if pubMsg, err := privatePlayersInfo.ToJSON(); err == nil {
			rdb.Publish(ctx, "room:"+r.ID+":broadcast", pubMsg).Err()
		}
	}

	return nil
}

// SubscribeRoomBroadcast assina o canal de broadcast de uma sala
func (r *Room) SubscribeRoomBroadcast(rdb *redis.Client) {
	if structs.ConnManager.IsSubscribed(r.ID) {
		utils.LogDebug("Sala " + r.ID + " já está assinada no Pub/Sub")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	structs.ConnManager.SetRoomCancel(r.ID, cancel)
	structs.ConnManager.MarkSubscribed(r.ID)

	pubsub := rdb.Subscribe(ctx, "room:"+r.ID+":broadcast")

	go func() {
		defer func() {
			_ = pubsub.Close()
			structs.ConnManager.ClearSubscribed(r.ID)
			utils.LogDebug("Pub/Sub encerrado da sala " + r.ID)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return
				}

				// Tenta parsear como BroadcastMessage
				broadcastMsg, err := structs.ParseBroadcastMessage([]byte(msg.Payload))
				if err != nil {
					// Se falhar, envia a mensagem original para manter compatibilidade
					utils.LogError(fmt.Errorf("erro ao parsear BroadcastMessage, enviando payload original: %w", err))
					structs.ConnManager.Broadcast(r.ID, []byte(msg.Payload))
					continue
				}

				// Serializa apenas os dados para enviar aos clientes
				dataPayload, err := json.Marshal(broadcastMsg.Data)
				if err != nil {
					utils.LogError(fmt.Errorf("erro ao serializar dados da BroadcastMessage: %w", err))
					continue
				}

				// Decide se é broadcast para todos ou seletivo
				if broadcastMsg.ToAll {
					// Envia para todos os jogadores conectados na sala
					structs.ConnManager.Broadcast(r.ID, dataPayload)
				} else {
					// Envia apenas para os jogadores especificados
					structs.ConnManager.BroadcastSelective(r.ID, dataPayload, broadcastMsg.ConfidencialPlayerIds)
				}
			}
		}
	}()

	utils.LogDebug("Assinada sala " + r.ID + " no Pub/Sub")
}
