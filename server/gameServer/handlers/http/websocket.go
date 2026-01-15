package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
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

		playerId := claims["playerId"].(string)
		username := claims["username"].(string)

		roomDataConn, err := redisFuncs.LoadRoom(ctxConn, rdb, claims["roomId"].(string))
		if err != nil {
			utils.LogError("Erro ao carregar dados da sala: " + err.Error())
			conn.Close()
			return
		}

		// Assina canal Redis Pub/Sub da sala
		roomDataConn.SubscribeRoomBroadcast(ctxConn, rdb)

		// Adiciona conexão ao ConnManager
		roomStructs.ConnManager.Add(roomDataConn.ID, playerId, conn)
		defer roomStructs.ConnManager.Disconnect(roomDataConn.ID, playerId)

		// Registra que player está em uma sala no Redis
		roomDataConn.RegisterPlayerInRoom(ctxConn, rdb, playerId)
		defer playerRedis.UnregisterPlayerFromRoom(ctxConn, rdb, playerId)

		// Chamadas de hook
		connection.OnConnect(conn, ctxConn, rdb, roomDataConn, playerId, username)
		defer connection.OnDisconnect(conn, ctxConn, rdb, playerId, roomDataConn)

		player, err := roomDataConn.GetPlayer(playerId)
		if err != nil || player == nil {
			utils.LogError("Erro ao carregar dados do player na sala: " + err.Error())
			conn.Close()
			return
		}

		// Loop de leitura das mensagens do cliente
		for {
			ctx := r.Context()

			_, msg, err := conn.ReadMessage()
			if err != nil {
				utils.LogError(fmt.Sprintf("Erro ao ler mensagem: %v", err))
				break
			}
			utils.LogDebug(fmt.Sprintf("Mensagem detectada do player %s na sala %s", player.Id, roomDataConn.ID))

			roomId, instanceId, registered, err := player.GetPlayerRegistrationInfo(ctx, rdb)
			if err != nil || !registered || roomId == "" || instanceId == "" {
				utils.LogError("Conexão perdida ou não registrada")
				conn.Close()
				break
			}

			// Valida e decodifica a jogada do player
			play, err := roomStructs.ParsePlayerPlay(msg, player.Id)
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
