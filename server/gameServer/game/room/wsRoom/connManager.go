package wsRoom

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	Conn *websocket.Conn
	mu   sync.Mutex // protege writes/close na conexão
}

type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]map[string]*WSConn // roomId -> playerId -> WSConn
	// subscriptions guarda se esta instância está inscrita no broadcast de uma room
	subscriptions map[string]bool // roomId -> subscribed
}

var ConnManager = &ConnectionManager{
	connections:   make(map[string]map[string]*WSConn),
	subscriptions: make(map[string]bool),
}

func (cm *ConnectionManager) Add(roomId, playerId string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.connections[roomId] == nil {
		cm.connections[roomId] = make(map[string]*WSConn)
	}
	cm.connections[roomId][playerId] = &WSConn{Conn: conn}
}

// Disconnect remove a conexão do mapa e fecha a conexão de forma segura.
// Se não houver mais conexões para a sala, realiza unsubscribe do broadcast.
func (cm *ConnectionManager) Disconnect(roomId, playerId string) {
	// Remover referência do mapa sob lock, mas fazer Close() fora do lock
	var ws *WSConn
	unsubscribe := false

	cm.mu.Lock()
	if cm.connections[roomId] != nil {
		if w, exists := cm.connections[roomId][playerId]; exists {
			ws = w
			delete(cm.connections[roomId], playerId)
		}
		if cm.connections[roomId] != nil && len(cm.connections[roomId]) == 0 {
			delete(cm.connections, roomId)
			if cm.subscriptions[roomId] {
				delete(cm.subscriptions, roomId)
				unsubscribe = true
			}
		}
	}
	cm.mu.Unlock()

	if ws != nil {
		ws.mu.Lock()
		ws.Conn.Close()
		ws.mu.Unlock()
	}

	if unsubscribe {
		UnsubscribeRoomBroadcast(roomId)
	}
}

// Broadcast envia mensagem para todas as conexões locais da sala.
// Protege cada conn com seu próprio mutex para evitar concurrent writes.
func (cm *ConnectionManager) Broadcast(roomId string, message []byte) {
	// Copia as referências sob read lock para evitar segurar o lock durante writes
	cm.mu.RLock()
	var conns []*WSConn
	if players, ok := cm.connections[roomId]; ok {
		conns = make([]*WSConn, 0, len(players))
		for _, ws := range players {
			conns = append(conns, ws)
		}
	}
	cm.mu.RUnlock()

	for _, ws := range conns {
		ws.mu.Lock()
		_ = ws.Conn.WriteMessage(websocket.TextMessage, message)
		ws.mu.Unlock()
	}
}

// IsSubscribed retorna true se esta instância já está inscrita no broadcast da room
func (cm *ConnectionManager) IsSubscribed(roomId string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.subscriptions[roomId]
}

// MarkSubscribed marca explicitamente que esta instância está inscrita no broadcast
func (cm *ConnectionManager) MarkSubscribed(roomId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.subscriptions[roomId] = true
}

// MarkUnsubscribed remove a flag de subscription (usado quando fizermos unsubscribe do pubsub)
func (cm *ConnectionManager) MarkUnsubscribed(roomId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.subscriptions, roomId)
}
