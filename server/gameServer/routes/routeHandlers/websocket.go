package routeHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/routeHandlers/wsHandlers"
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

type WebSocketPayload struct {
	Type     string `json:"type"`
	PlayerId string `json:"playerId"`
	RoomId   string `json:"roomId"`
	Payload  any    `json:"payload"`
}

// WebSocketHandler lida com conexões WS
// Enviar Token JWT na url: /game?ticket=SEU_TOKEN_JWT
func (h *Handler) WebSocketHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Valida e faz upgrade para WebSocket
		conn, claims := wsHandlers.ValidateConnection(w, r, h.JWTService, h.RoomService, upgrader, rdb, ctx)
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
		redisHandlers.RegisterPlayerInRoom(ctx, rdb, playerId, roomId)
		defer redisHandlers.UnregisterPlayerFromRoom(ctx, rdb, playerId)

		// Chamadas de hook
		wsHandlers.OnConnect(conn, ctx, rdb, roomId, playerId, username)
		defer wsHandlers.OnDisconnect(conn, ctx, rdb, playerId, roomId)

		// Loop de leitura das mensagens do cliente
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				utils.LogDebug(fmt.Sprintf("Erro ao ler mensagem: %v", err))
				break
			}
			utils.LogDebug(fmt.Sprintf("Mensagem detectada do player %s na sala %s", playerId, roomId))

			roomId, instanceId, registered, err := redisHandlers.GetPlayerRegistrationInfo(ctx, rdb, playerId)
			if err != nil || !registered || roomId == "" || instanceId == "" {
				utils.LogDebug("Conexão perdida ou não registrada")
				conn.Close()
				break
			}

			// Decodifica evento
			var playerPlay roomStructs.PendingEvent
			if err := json.Unmarshal(msg, &playerPlay); err != nil {
				utils.LogDebug(fmt.Sprintf("Erro ao decodificar mensagem: %v", err))
				break
			}
			playerPlay.PlayerID = playerId

			// Processa mensagem normalmente
			wsHandlers.OnMessage(ctx, conn, rdb, &playerPlay, roomId)
		}
	}
}
