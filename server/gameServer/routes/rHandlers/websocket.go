package rHandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	rs "github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/rHandlers/wsHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == os.Getenv("FRONTEND_URL") || origin == ""
	},
}

type WebSocketPayload struct {
	Type     string      `json:"type"`
	PlayerId string      `json:"playerId"`
	RoomId   string      `json:"roomId"`
	Payload  interface{} `json:"payload"`
}

// WebSocketHandler lida com conexões WS
func (h *Handler) WebSocketHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		conn, claims, err := wsHandlers.ValidateConnection(w, r, h.JWTService, upgrader)
		if err != nil {
			utils.SendError(w, "WebSocket upgrade failed", http.StatusInternalServerError)
			return
		}

		defer func() {
			wsHandlers.OnDisconnect(conn)
			conn.Close()
		}()

		wsHandlers.OnConnect(conn)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				utils.LogDebug(fmt.Sprintf("Erro ao ler mensagem: %v", err))
				break
			}

			var event rs.Event
			if err := json.Unmarshal(msg, &event); err != nil {
				utils.LogDebug(fmt.Sprintf("Erro ao decodificar mensagem: %v", err))
				break
			}
			wsHandlers.OnMessage(ctx, conn, rdb, &event, claims)
		}
	}
}
