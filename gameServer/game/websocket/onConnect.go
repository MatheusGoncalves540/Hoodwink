package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

// Função chamada quando o cliente se conecta ao WebSocket
func OnConnect(conn *websocket.Conn) {
	log.Println("Cliente conectado ao WebSocket")
	// Aqui você pode adicionar lógica extra, como enviar mensagem de boas-vindas
}
