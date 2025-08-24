package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// Valida a conexão WebSocket e retorna o objeto de conexão ou um erro
func ValidateConnection(w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return nil, err
	}
	return conn, nil
}

// Função para logar erros comuns
func LogError(context string, err error) {
	if err != nil {
		log.Println(context+":", err)
	}
}
