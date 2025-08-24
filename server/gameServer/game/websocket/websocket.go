package websocket

import (
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/websocket/auth"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // ajuste para produção
}

type WebSocketPayload struct {
	Type     string      `json:"type"`
	PlayerId string      `json:"playerId"`
	RoomId   string      `json:"roomId"`
	Payload  interface{} `json:"payload"`
}

// WebSocketHandler lida com conexões WS
func WebSocketHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		claims, err := auth.ParseTokenFromRequest(r)
		if err != nil {
			utils.SendError(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		playerID, _ := claims["playerId"].(string)
		roomID, _ := claims["roomId"].(string)
		fmt.Println("Usuário autenticado via JWT:", playerID, "na sala", roomID)

		conn, err := ValidateConnection(w, r, upgrader)
		if err != nil {
			utils.SendError(w, "WebSocket upgrade failed", http.StatusInternalServerError)
			return
		}
		defer func() {
			OnDisconnect(conn)
			conn.Close()
		}()

		OnConnect(conn)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				LogError("Erro ao ler mensagem", err)
				break
			}

			OnMessage(ctx, conn, rdb, msg)
		}
	}
}
