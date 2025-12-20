package wsHandlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/wsRoom"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Valida a conexão WebSocket e retorna o objeto de conexão ou um erro
func ValidateConnection(w http.ResponseWriter, r *http.Request, jwtService *services.JWTService, roomService *services.RoomService, upgrader websocket.Upgrader, rdb *redis.Client, ctx context.Context) (*websocket.Conn, jwt.MapClaims) {
	claims, err := jwtService.ParseTokenFromRequest(r)
	if err != nil {
		utils.SendError(w, "Ticket invalido: "+err.Error(), http.StatusUnauthorized)
		return nil, nil
	}

	playerID, _ := claims["playerId"].(string)
	roomID, _ := claims["roomId"].(string)

	// Sincroniza jogadores ativos na sala antes de permitir a conexão se o jogo não tiver iniciado
	if startTimeAny, err := redisHandlers.LoadRoomField(ctx, rdb, roomID, "StartTime"); err != nil {
		utils.LogDebug("Erro ao verificar horário de início da sala:" + err.Error())
		return nil, nil
	} else if startTime, ok := startTimeAny.(time.Time); !ok {
		utils.LogDebug("Erro: StartTime não é do tipo time.Time")
		return nil, nil
	} else if startTime.IsZero() {
		// Jogo não iniciado, sincroniza a estrutura dos players na sala
		if err := roomService.SyncActivePlayers(r, rdb, roomID); err != nil {
			utils.LogDebug("Erro ao sincronizar jogadores da estrutura da sala: " + err.Error())
		}
	}

	connectedToRoomID, registered, err := redisHandlers.GetRegisteredRoomForPlayer(ctx, rdb, playerID)
	if err != nil {
		utils.SendError(w, "Falha ao obter sala registrada", http.StatusInternalServerError)
		utils.LogDebug("Falha ao obter sala registrada: " + err.Error())
		return nil, nil
	}

	if registered && connectedToRoomID != roomID {
		//Se já está registrado em outra sala que não é essa que está tentando entrar
		utils.SendError(w, "Player já está conectado em outra sala: "+connectedToRoomID, http.StatusConflict)
		return nil, nil
	} else if registered && connectedToRoomID != "" {
		// Se já está registrado na mesma sala, força reconexão: desconecta a anterior
		wsRoom.ConnManager.Disconnect(roomID, playerID)
	}

	_, err = redisHandlers.LoadRoom(r.Context(), rdb, roomID)
	if err != nil {
		utils.SendError(w, "Sala não encontrada", http.StatusNotFound)
		return nil, nil
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.SendError(w, "Erro ao verificar sala na entrada", http.StatusInternalServerError)
		utils.LogDebug(fmt.Sprintf("Verify FRONTEND_URL, WebSocket upgrade error: %v", err))
		return nil, nil
	}

	utils.LogDebug(fmt.Sprintf("Usuário autenticado via JWT: %s na sala %s", playerID, roomID))

	return conn, claims
}
