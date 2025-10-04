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
func (h *Handler) WebSocketHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Valida e faz upgrade para WebSocket
		conn, claims := wsHandlers.ValidateConnection(w, r, h.JWTService, upgrader, rdb, ctx)
		if conn == nil || claims == nil {
			return
		}

		playerId := claims["playerId"].(string)
		roomId := claims["roomId"].(string)

		// Adiciona conexão ao ConnManager
		wsRoom.ConnManager.Add(roomId, playerId, conn)
		defer wsRoom.ConnManager.Remove(roomId, playerId)

		// Assina canal Redis Pub/Sub da sala
		wsRoom.SubscribeRoomBroadcast(ctx, rdb, roomId)

		// Registra que player está em uma sala no Redis
		redisHandlers.RegisterPlayerInRoom(ctx, rdb, playerId, roomId)

		// Chamadas de hook
		wsHandlers.OnConnect(conn, ctx, rdb, roomId, playerId)
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
			var event roomStructs.Event
			if err := json.Unmarshal(msg, &event); err != nil {
				utils.LogDebug(fmt.Sprintf("Erro ao decodificar mensagem: %v", err))
				break
			}

			// Processa mensagem normalmente
			wsHandlers.OnMessage(ctx, conn, rdb, &event, claims)

			// 4️⃣ Publica o evento para outras instâncias
			if pubMsg, err := json.Marshal(event); err == nil {
				wsRoom.PublishRoomBroadcast(ctx, rdb, roomId, pubMsg)
			}
		}
	}
}
