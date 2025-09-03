package wsHandlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// Valida a conexão WebSocket e retorna o objeto de conexão ou um erro
func ValidateConnection(w http.ResponseWriter, r *http.Request, jwtService *services.JWTService, upgrader websocket.Upgrader) (*websocket.Conn, jwt.MapClaims, error) {
	claims, err := jwtService.ParseTokenFromRequest(r)
	if err != nil {
		utils.SendError(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return nil, nil, err
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return nil, nil, err
	}

	playerID, _ := claims["playerId"].(string)
	roomID, _ := claims["roomId"].(string)
	utils.LogDebug(fmt.Sprintf("Usuário autenticado via JWT: %s na sala %s", playerID, roomID))

	return conn, claims, nil
}
