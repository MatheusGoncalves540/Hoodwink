package connection

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada quando o cliente se conecta ao WebSocket
func OnConnect(conn *websocket.Conn, ctx context.Context, rdb *redis.Client, rulesRegistry *rules.Registry, roomId string, playerId string, username string) {
	// reconexão no meio da partida não pode gravar uma sala velha por cima da jogada em curso
	err := redisFuncs.WithRoom(ctx, rdb, roomId, redisFuncs.CauseConnect, func(roomData *rooms.Room) bool {
		aplicado, err := redisFuncs.AddPlayerInRoom(rulesRegistry, roomData, playerId, username)
		if err != nil {
			utils.LogError("Erro ao adicionar/atualizar jogador na sala: " + err.Error())
			return false
		}
		utils.LogDebug("Cliente conectado ao WebSocket")
		redisFuncs.SendUpdatedRoomData(ctx, rdb, roomData, nil, []string{})
		return aplicado
	})
	if err != nil {
		utils.LogError("Erro ao carregar dados da sala no connect: " + err.Error())
	}
}
