package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Função chamada ao receber uma mensagem do cliente
func OnMessage(ctx context.Context, conn *websocket.Conn, rdb *redis.Client, msg []byte) {
	var evt WebSocketPayload
	if err := json.Unmarshal(msg, &evt); err != nil {
		log.Println("Formato inválido:", err)
		return
	}

	roomId, err := redisHandlers.LoadRoomField(ctx, rdb, evt.RoomId, "ID")
	if err != nil {
		log.Println("Erro ao buscar sala:", err)
		return
	}

	fmt.Print(roomId)

	// ScheduleNextStep dentro da função de mensagem
	// redisHandlers.ScheduleNextStep(ctx, rdb, roomId, eventQueue.Event{
	// 	Type:      "no_contest",
	// 	PlayerId:  "system",
	// 	TimeoutMs: 8000,
	// })

	// Atualize todos os jogadores (exemplo simplificado)
	// stateBytes, _ := json.Marshal(roomData)
	// conn.WriteMessage(websocket.TextMessage, stateBytes)
}
