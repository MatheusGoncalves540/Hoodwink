package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/connection"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/player"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		if os.Getenv("CORS") == "false" {
			return true
		}
		origin := r.Header.Get("Origin")
		return origin == os.Getenv("FRONTEND_URL") || origin == ""
	},
}

// WebSocketHandler lida com conexões WS
// Enviar Token JWT na url: /game?ticket=SEU_TOKEN_JWT
func (h *HTTPHandler) WebSocketHandler(rdb *redis.Client, rulesRegistry *rules.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Valida e faz upgrade para WebSocket
		conn, claims := connection.ValidateConnection(w, r, h.JWTService, h.RoomService, upgrader, rdb, ctx)
		if conn == nil || claims == nil {
			return
		}

		playerId := claims["playerId"].(string)
		username := claims["username"].(string)
		roomId := claims["roomId"].(string)

		// Assina canal Redis Pub/Sub da sala
		wsRoom.SubscribeRoomBroadcast(ctx, rdb, roomId)

		// Adiciona conexão ao ConnManager
		wsRoom.ConnManager.Add(roomId, playerId, conn)
		defer wsRoom.ConnManager.Disconnect(roomId, playerId)

		// Registra que player está em uma sala no Redis
		player.RegisterPlayerInRoom(ctx, rdb, playerId, roomId)
		defer player.UnregisterPlayerFromRoom(ctx, rdb, playerId)

		// Chamadas de hook
		connection.OnConnect(conn, ctx, rdb, roomId, playerId, username)
		defer connection.OnDisconnect(conn, ctx, rdb, playerId, roomId)

		// Loop de leitura das mensagens do cliente
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				utils.LogError(fmt.Sprintf("Erro ao ler mensagem: %v", err))
				break
			}
			utils.LogDebug(fmt.Sprintf("Mensagem detectada do player %s na sala %s", playerId, roomId))

			roomId, instanceId, registered, err := player.GetPlayerRegistrationInfo(ctx, rdb, playerId)
			if err != nil || !registered || roomId == "" || instanceId == "" {
				utils.LogError("Conexão perdida ou não registrada")
				conn.Close()
				break
			}

			// Valida e decodifica a jogada do player
			play, err := roomStructs.ParsePlayerPlay(msg, playerId)
			if err != nil {
				utils.LogError(fmt.Sprintf("Evento inválido: %v", err))
				conn.Close()
				break
			}

			connection.OnMessage(ctx, conn, rdb, rulesRegistry, play, roomId)
		}
	}
}
