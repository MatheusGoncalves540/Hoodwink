package game

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/eventQueue"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // ajuste para produção
}

type WebSocketPayload struct {
	Type      string          `json:"type"`
	PlayerId  string          `json:"playerId"`
	RoomId    string          `json:"roomId"`
	Payload   json.RawMessage `json:"payload"`
	TimeoutMs int             `json:"timeoutMs,omitempty"` // opcional
}

// WebSocketHandler lida com conexões WS
func WebSocketHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Erro ao ler mensagem:", err)
				break
			}

			var evt WebSocketPayload
			if err := json.Unmarshal(msg, &evt); err != nil {
				log.Println("Formato inválido:", err)
				continue
			}

			roomData, err := redisHandlers.LoadRoom(ctx, rdb, evt.RoomId)
			if err != nil {
				log.Println("Erro ao buscar sala:", err)
				continue
			}

			event := eventQueue.Event{
				Type:      evt.Type,
				PlayerId:  evt.PlayerId,
				RoomId:    evt.RoomId,
				Payload:   evt.Payload,
				TimeoutMs: evt.TimeoutMs,
				CreatedAt: time.Now(),
			}

			err = room.ProcessEvent(ctx, rdb, roomData, &event)
			if err != nil {
				log.Println("Erro no ProcessEvent:", err)
				continue
			}

			// Atualize todos os jogadores (exemplo simplificado)
			stateBytes, _ := json.Marshal(roomData)
			conn.WriteMessage(websocket.TextMessage, stateBytes)
		}
	}
}
