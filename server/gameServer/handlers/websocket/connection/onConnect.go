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
	err := redisFuncs.WithRoom(ctx, rdb, roomId, func(roomData *rooms.Room) {
		if err := roomData.AddPlayerInRoom(ctx, rdb, rulesRegistry, playerId, username); err != nil {
			utils.LogError("Erro ao adicionar/atualizar jogador na sala: " + err.Error())
		}
		utils.LogDebug("Cliente conectado ao WebSocket")
		roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{})
	})
	if err != nil {
		utils.LogError("Erro ao carregar dados da sala no connect: " + err.Error())
	}
}
