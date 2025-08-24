package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

// Função chamada quando o cliente se desconecta do WebSocket
func OnDisconnect(conn *websocket.Conn) {
	log.Println("Cliente desconectado do WebSocket")
	// Aqui você pode adicionar lógica extra, como limpar recursos
}
