package redisFuncs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/connmanager"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// BroadcastMsgToRoom publica mensagem para todos os players de uma sala.
func BroadcastMsgToRoom(ctx context.Context, rdb *redis.Client, room *rooms.Room, message any) error {
	if pubMsg, err := json.Marshal(message); err == nil {
		return rdb.Publish(ctx, "room:"+room.ID+":broadcast", pubMsg).Err()
	}
	return nil
}

// SendUpdatedRoomData publica nova versão da sala para todos os players ou para players específicos.
// Se confidencialRoomData for nil OU confidencialPlayerIds estiver vazio, envia para todos (toAll: true).
// Caso contrário, envia apenas para os playerIds especificados usando confidencialRoomData.
func SendUpdatedRoomData(ctx context.Context, rdb *redis.Client, room *rooms.Room, confidencialRoomData *rooms.Room, confidencialPlayerIds []string) error {
	if rdb == nil {
		return nil
	}

	toAll := confidencialRoomData == nil || len(confidencialPlayerIds) == 0

	if toAll {
		if utils.IsUnsafeDebugMode() {
			return PublishUnsafeRoomData(ctx, rdb, room, true, nil)
		}
		return PublishRoomUpdate(ctx, rdb, room, true, nil)
	}

	if utils.IsUnsafeDebugMode() {
		PublishUnsafeRoomData(ctx, rdb, room, true, nil)
		return PublishUnsafeRoomData(ctx, rdb, confidencialRoomData, false, confidencialPlayerIds)
	}

	PublishRoomUpdate(ctx, rdb, room, true, nil)
	return PublishRoomUpdate(ctx, rdb, confidencialRoomData, false, confidencialPlayerIds)
}

// PublishUnsafeRoomData publica a versão privada completa, não segura da sala (APENAS PARA DEBUG).
func PublishUnsafeRoomData(ctx context.Context, rdb *redis.Client, room *rooms.Room, toAll bool, playerIds []string) error {
	var broadcastMsg *structs.BroadcastMessage

	if toAll {
		broadcastMsg = structs.NewBroadcastMessage(room)
	} else {
		broadcastMsg = structs.NewSelectiveBroadcastMessage(room, playerIds)
	}

	if pubMsg, err := broadcastMsg.ToJSON(); err == nil {
		return rdb.Publish(ctx, "room:"+room.ID+":broadcast", pubMsg).Err()
	}
	return nil
}

// PublishRoomUpdate publica a versão pública da sala para todos os players de uma sala.
// No caminho toAll publica primeiro o payload público (com Chat) e depois a versão
// privada de cada jogador (sem Chat, só com as cartas da mão dele).
func PublishRoomUpdate(ctx context.Context, rdb *redis.Client, room *rooms.Room, toAll bool, playerIds []string) error {
	roomDataPublic := room.GetPublicRoomForUpdates()

	if toAll {
		if pubMsg, err := structs.NewBroadcastMessage(roomDataPublic).ToJSON(); err == nil {
			if err := rdb.Publish(ctx, "room:"+room.ID+":broadcast", pubMsg).Err(); err != nil {
				return err
			}
		}
	} else {
		broadcastMsg := structs.NewSelectiveBroadcastMessage(roomDataPublic, playerIds)
		if pubMsg, err := broadcastMsg.ToJSON(); err == nil {
			return rdb.Publish(ctx, "room:"+room.ID+":broadcast", pubMsg).Err()
		}
	}

	// Sempre que for enviar um broadcast, mesmo que seja para todos, envia também a versão privada para cada player individualmente
	// para garantir que cada player receba os dados confidenciais que só ele pode ver (como as cartas da mão).
	for _, player := range room.Players {
		// O chat já vai no broadcast público acima; o privado só precisa dos dados
		// confidenciais do jogador, sem carregar (nem clonar) o histórico de chat.
		roomDataPrivate := clonePublicForPrivate(roomDataPublic)
		roomDataPrivate.Players[player.Id] = player.GetPrivatePlayerForPublicUpdates()

		privatePlayersInfo := structs.NewSelectiveBroadcastMessage(roomDataPrivate, []string{player.Id})

		if pubMsg, err := privatePlayersInfo.ToJSON(); err == nil {
			rdb.Publish(ctx, "room:"+room.ID+":broadcast", pubMsg).Err()
		}
	}

	return nil
}

// clonePublicForPrivate copia a sala pública para o envio privado de um jogador sem
// serializar/desserializar tudo de novo: o Chat vai só no broadcast público e o
// baralho já é apenas a contagem (Deck int). Os valores do mapa são structs, então
// basta copiar o mapa: só a entrada do jogador dono do payload é substituída.
func clonePublicForPrivate(public rooms.PublicRoomForUpdates) *rooms.PublicRoomForUpdates {
	clone := public
	clone.Chat = nil
	clone.Players = make(map[string]players.PublicPlayerForUpdates, len(public.Players))
	for playerId, player := range public.Players {
		clone.Players[playerId] = player
	}
	return &clone
}

// SubscribeRoomBroadcast assina o canal de broadcast de uma sala.
func SubscribeRoomBroadcast(rdb *redis.Client, room *rooms.Room) {
	if connmanager.ConnManager.IsSubscribed(room.ID) {
		utils.LogDebug("Sala " + room.ID + " já está assinada no Pub/Sub")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	connmanager.ConnManager.SetRoomCancel(room.ID, cancel)
	connmanager.ConnManager.MarkSubscribed(room.ID)

	pubsub := rdb.Subscribe(ctx, "room:"+room.ID+":broadcast")

	go func() {
		defer func() {
			_ = pubsub.Close()
			connmanager.ConnManager.ClearSubscribed(room.ID)
			utils.LogDebug("Pub/Sub encerrado da sala " + room.ID)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return
				}

				broadcastMsg, err := structs.ParseBroadcastMessage([]byte(msg.Payload))
				if err != nil {
					utils.LogError(fmt.Errorf("erro ao parsear BroadcastMessage, enviando payload original: %w", err))
					connmanager.ConnManager.Broadcast(room.ID, []byte(msg.Payload))
					continue
				}

				dataPayload, err := json.Marshal(broadcastMsg.Data)
				if err != nil {
					utils.LogError(fmt.Errorf("erro ao serializar dados da BroadcastMessage: %w", err))
					continue
				}

				if broadcastMsg.ToAll {
					connmanager.ConnManager.Broadcast(room.ID, dataPayload)
				} else {
					connmanager.ConnManager.BroadcastSelective(room.ID, dataPayload, broadcastMsg.ConfidencialPlayerIds)
				}
			}
		}
	}()

	utils.LogDebug("Assinada sala " + room.ID + " no Pub/Sub")
}
