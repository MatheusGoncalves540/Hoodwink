package connection

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Valida a conexão WebSocket e retorna o objeto de conexão ou um erro
func ValidateConnection(w http.ResponseWriter, r *http.Request, jwtService *services.JWTService, roomService *services.RoomService, upgrader websocket.Upgrader, rdb *redis.Client, ctx context.Context) (*websocket.Conn, jwt.MapClaims) {
	// Valida o token JWT e extrai as claims
	claims, err := jwtService.ParseTokenFromRequest(r)
	if err != nil {
		utils.SendError(w, "Ticket invalido: "+err.Error(), http.StatusUnauthorized)
		return nil, nil
	}

	playerID, okPlayer := claims["playerId"].(string)
	roomID, okRoom := claims["roomId"].(string)
	username, okUsername := claims["username"].(string)
	if !okPlayer || playerID == "" || !okRoom || roomID == "" || !okUsername || username == "" {
		utils.SendError(w, "Ticket inválido: claims ausentes", http.StatusUnauthorized)
		return nil, nil
	}

	// Carrega os dados da sala
	roomData, err := redisFuncs.LoadRoom(ctx, rdb, roomID)
	if err != nil {
		utils.SendError(w, "Sala não encontrada", http.StatusNotFound)
		utils.LogError("Erro ao verificar informações da sala: " + err.Error())
		return nil, nil
	}

	// Jogo não iniciado, sincroniza a estrutura dos players na sala
	if roomData.StartTime.IsZero() {
		synced, err := roomService.SyncActivePlayers(r, rdb, roomData)
		if err != nil {
			utils.LogError("Erro ao sincronizar jogadores da estrutura da sala: " + err.Error())
		} else {
			roomData = synced
		}
	}

	// Revalida a associação jogador↔sala no upgrade (regra local equivalente ao
	// ValidatePlayerEntry): jogador novo não entra em sala cheia nem em jogo já iniciado.
	player, _ := roomData.GetPlayer(playerID)
	if player == nil {
		if roomData.IsFull() {
			utils.SendError(w, "Sala está cheia", http.StatusUnauthorized)
			return nil, nil
		}
		if !roomData.StartTime.IsZero() {
			utils.SendError(w, "Jogo já começou", http.StatusUnauthorized)
			return nil, nil
		}
	}

	// Verifica se o player já está registrado em uma sala
	connectedToRoomID, registered, err := playerRedis.GetRegisteredRoomForPlayer(ctx, rdb, playerID)
	if err != nil {
		utils.SendError(w, "Falha ao obter sala registrada", http.StatusInternalServerError)
		utils.LogError("Falha ao obter sala registrada: " + err.Error())
		return nil, nil
	}

	if registered {
		//Se já está registrado em outra sala que não é essa que está tentando entrar
		utils.SendError(w, "Player já está conectado em uma sala: "+connectedToRoomID, http.StatusConflict)
		return nil, nil
	}

	// Faz upgrade ANTES de qualquer validação que possa falhar após hijack
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogError(fmt.Sprintf("Verify FRONTEND_URL, WebSocket upgrade error: %v", err))
		return nil, nil
	}

	utils.LogDebug(fmt.Sprintf("Usuário autenticado via JWT: %s na sala %s", playerID, roomID))

	return conn, claims
}
