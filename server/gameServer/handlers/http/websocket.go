package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/connection"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Configuração do upgrader WebSocket
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
		ctxConn := r.Context()

		// Valida e faz upgrade para WebSocket
		conn, claims := connection.ValidateConnection(w, r, h.JWTService, h.RoomService, upgrader, rdb, ctxConn)
		if conn == nil || claims == nil {
			return
		}

		roomId := claims["roomId"].(string)
		playerId := claims["playerId"].(string)
		username := claims["username"].(string)

		roomObj, err := redisFuncs.LoadRoom(ctxConn, rdb, roomId)
		if err != nil {
			utils.LogError("Erro ao carregar dados da sala: " + err.Error())
			conn.Close()
			return
		}

		// Assina canal Redis Pub/Sub da sala
		roomObj.SubscribeRoomBroadcast(rdb)

		// Adiciona conexão ao ConnManager
		structs.ConnManager.Add(roomId, playerId, conn)
		defer structs.ConnManager.Disconnect(roomId, playerId)

		// Registra que player está em uma sala no Redis
		playerRedis.RegisterPlayerInRoom(ctxConn, rdb, roomId, playerId)
		defer playerRedis.UnregisterPlayerFromRoom(ctxConn, rdb, playerId)

		// Chamadas de hook
		connection.OnConnect(conn, ctxConn, rdb, roomId, playerId, username)
		defer connection.OnDisconnect(conn, ctxConn, rdb, roomId, playerId)

		// Loop de leitura das mensagens do cliente
		for {
			ctx := r.Context()

			_, msg, err := conn.ReadMessage()
			if err != nil {
				utils.LogError(fmt.Sprintf("Erro ao ler mensagem: %v", err))
				break
			}

			utils.LogDebug(fmt.Sprintf("Mensagem detectada do player %s na sala %s", playerId, roomObj.ID))

			roomId, instanceId, registered, err := playerRedis.GetPlayerRegistrationInfo(ctx, rdb, playerId)
			if err != nil || !registered || roomId == "" || instanceId == "" {
				utils.LogError("Conexão perdida ou não registrada")
				conn.Close()
				break
			}

			// Valida e decodifica a jogada do player
			play, err := structs.ParsePlayerPlay(msg, playerId)
			if err != nil {
				utils.LogError(fmt.Sprintf("Evento inválido: %v", err))
				conn.Close()
				break
			}

			roomData, err := redisFuncs.LoadRoom(ctx, rdb, roomId)
			if err != nil {
				utils.LogError(fmt.Sprintf("Erro ao carregar dados da sala: %v", err))
				conn.Close()
				break
			}

			connection.OnMessage(ctx, rdb, rulesRegistry, play, roomData)
			// Salva e publica atualizações da sala
			roomData.SaveRoom(ctx, rdb)
			roomData.PublishRoomBroadcast(ctx, rdb, roomData)
		}
	}
}
