package connection

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se desconecta do WebSocket
func OnDisconnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, roomId string, playerId string) {
	playerRedis.AcquirePlayerLock(ctx, rdb, playerId, 0)
	defer playerRedis.ReleasePlayerLock(ctx, rdb, playerId)

	err := redisFuncs.WithRoom(ctx, rdb, roomId, redisFuncs.CauseDisconnect, func(roomData *rooms.Room) bool {
		return onDisconnect(ctx, rdb, roomData, playerId)
	})
	if err != nil {
		utils.LogError("Erro ao carregar dados da sala no disconnect: " + err.Error())
	}
}

func onDisconnect(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerId string) bool {
	aplicado := false

	// Verifica se o jogo já foi iniciado
	if roomData.StartTime.IsZero() {
		// Jogo não iniciado, remove a estrutura do player na sala
		removed, err := redisFuncs.RemovePlayerFromRoom(roomData, playerId)
		if err != nil {
			utils.LogError("Erro ao remover jogador da estrutura da sala: " + err.Error())
		} else {
			aplicado = removed
		}
	} else {
		// Jogo iniciado: o jogador permanece na sala mas desconecta.
		roomData.RemoveConnected(playerId)
		aplicado = true
	}

	// Verifica se a sala está vazia após a desconexão
	if len(roomData.Connected) == 0 {
		err := redisFuncs.SetTTL(ctx, rdb, roomData, 20*time.Second)
		if err != nil {
			utils.LogError("Erro ao definir TTL da sala vazia: " + err.Error())
		} else {
			utils.LogDebug("TTL de 20 segundos definido para sala vazia " + roomData.ID)
		}
	}

	utils.LogDebug("Cliente desconectado do WebSocket")
	redisFuncs.SendUpdatedRoomData(ctx, rdb, roomData, nil, []string{})
	return aplicado
}
