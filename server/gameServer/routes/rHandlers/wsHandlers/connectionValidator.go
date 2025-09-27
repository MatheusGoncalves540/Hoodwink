package wsHandlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Valida a conexão WebSocket e retorna o objeto de conexão ou um erro
func ValidateConnection(w http.ResponseWriter, r *http.Request, jwtService *services.JWTService, upgrader websocket.Upgrader, rdb *redis.Client, ctx context.Context) (*websocket.Conn, jwt.MapClaims) {
	claims, err := jwtService.ParseTokenFromRequest(r)
	if err != nil {
		utils.SendError(w, "Ticket invalido: "+err.Error(), http.StatusUnauthorized)
		return nil, nil
	}

	playerID, _ := claims["playerId"].(string)
	roomID, _ := claims["roomId"].(string)

	connectedToRoomID, registered, err := redisHandlers.GetRegisteredRoomForPlayer(ctx, rdb, playerID)
	if err != nil {
		utils.SendError(w, "Falha ao obter sala registrada", http.StatusInternalServerError)
		utils.LogDebug("Falha ao obter sala registrada: " + err.Error())
		return nil, nil
	}

	if registered && connectedToRoomID != roomID {
		//Se já está registrado em outra sala que não é essa que está tentando entrar
		utils.SendError(w, "Player já está conectado em uma sala: "+connectedToRoomID, http.StatusConflict)
		return nil, nil
	} else if registered && connectedToRoomID != "" {
		//Verifica se o player está na sala e se está marcado como conectado
		room, err := redisHandlers.LoadRoom(ctx, rdb, connectedToRoomID)
		if err != nil {
			utils.SendError(w, "Erro ao verificar sala na entrada", http.StatusInternalServerError)
			utils.LogDebug("Erro ao verificar sala na entrada: " + err.Error())
			return nil, nil
		}

		player := utils.GetPlayerByID(room, playerID)
		if player == nil {
			utils.SendError(w, "Player não encontrado na sala que deveria estar conectado", http.StatusNotFound)
			return nil, nil
		}
		if player.Connected {
			utils.SendError(w, "Player já está conectado na sala", http.StatusForbidden)
			return nil, nil
		}
	}

	_, err = redisHandlers.LoadRoom(r.Context(), rdb, roomID)
	if err != nil {
		utils.SendError(w, "Sala não encontrada", http.StatusNotFound)
		return nil, nil
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.SendError(w, "Erro ao verificar sala na entrada", http.StatusInternalServerError)
		fmt.Print("Verify FRONTEND_URL, WebSocket upgrade error: ", err)
		return nil, nil
	}

	utils.LogDebug(fmt.Sprintf("Usuário autenticado via JWT: %s na sala %s", playerID, roomID))

	return conn, claims
}
