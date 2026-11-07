package http

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/connection"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/connmanager"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Configuração do upgrader WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Aceita clientes sem Origin (não-browser) ou exatamente FRONTEND_URL.
		return origin == "" || origin == os.Getenv("FRONTEND_URL")
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

		roomId, ok := claims["roomId"].(string)
		if !ok || roomId == "" {
			utils.LogError("Ticket inválido: claim roomId ausente")
			conn.Close()
			return
		}
		playerId, ok := claims["playerId"].(string)
		if !ok || playerId == "" {
			utils.LogError("Ticket inválido: claim playerId ausente")
			conn.Close()
			return
		}
		username, ok := claims["username"].(string)
		if !ok || username == "" {
			utils.LogError("Ticket inválido: claim username ausente")
			conn.Close()
			return
		}

		roomObj, err := redisFuncs.LoadRoom(ctxConn, rdb, roomId)
		if err != nil {
			utils.LogError("Erro ao carregar dados da sala: " + err.Error())
			conn.Close()
			return
		}

		// Assina canal Redis Pub/Sub da sala
		redisFuncs.SubscribeRoomBroadcast(rdb, roomObj)

		// Adiciona conexão ao ConnManager
		connmanager.ConnManager.Add(roomId, playerId, conn)
		defer connmanager.ConnManager.Disconnect(roomId, playerId)

		// Aplica limites de leitura (8 KB) e deadline renovável por pong/mensagem.
		configureConn(conn)
		pingCtx, pingCancel := context.WithCancel(ctxConn)
		defer pingCancel()
		startPingLoop(pingCtx, roomId, playerId)
		defer connRateLimiter.Delete(playerId)

		// Registra que player está em uma sala no Redis
		playerRedis.RegisterPlayerInRoom(ctxConn, rdb, roomId, playerId)
		defer playerRedis.UnregisterPlayerFromRoom(ctxConn, rdb, playerId)

		// Chamadas de hook
		connection.OnConnect(conn, ctxConn, rdb, rulesRegistry, roomId, playerId, username)
		defer connection.OnDisconnect(conn, ctxConn, rdb, roomId, playerId)

		// Loop de leitura das mensagens do cliente
		for {
			ctx := r.Context()

			_, msg, err := conn.ReadMessage()
			if err != nil {
				utils.LogError(fmt.Sprintf("Erro ao ler mensagem: %v", err))
				break
			}

			// Renova o deadline de leitura a cada mensagem (também renovado no pong).
			if err := renewReadDeadline(conn); err != nil {
				break
			}

			// Rate limit por conexão: acima de 20 mensagens/5s fecha só esta conexão.
			if !connRateLimiter.Allow(playerId) {
				utils.LogError(fmt.Sprintf("Rate limit excedido na conexão do player %s", playerId))
				conn.Close()
				break
			}

			// utils.LogDebug(fmt.Sprintf("Mensagem detectada do player %s na sala %s", playerId, roomObj.ID))
			roomId, instanceId, registered, err := playerRedis.GetPlayerRegistrationInfo(ctx, rdb, playerId)
			if err != nil || !registered || roomId == "" || instanceId == "" {
				utils.LogError("Conexão perdida ou não registrada")
				conn.Close()
				break
			}

			// Valida e decodifica a jogada do player
			play, err := structs.ParsePlayerPlay(msg, playerId)
			if err != nil {
				utils.LogInvldPlyrReq(fmt.Sprintf("Evento inválido: %v", err), playerId)
				conn.Close()
				break
			}

			// carrega a sala só depois de ter a vez dela: nada de gravar por cima do gameProcessor
			err = redisFuncs.WithRoom(ctx, rdb, roomId, redisFuncs.CausePlay, func(roomData *rooms.Room) bool {
				return connection.OnMessage(ctx, rdb, rulesRegistry, play, roomData)
			}, redisFuncs.ReplayPlayHint{Actor: play.PlayerId, Payload: msg})
			if err != nil {
				utils.LogError(fmt.Sprintf("Erro ao carregar dados da sala: %v", err))
				conn.Close()
				break
			}
		}
	}
}
