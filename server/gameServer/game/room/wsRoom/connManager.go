package wsRoom

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]map[string]*websocket.Conn // roomId -> playerId -> conn
}

var ConnManager = &ConnectionManager{
	connections: make(map[string]map[string]*websocket.Conn),
}

func (cm *ConnectionManager) Add(roomId, playerId string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.connections[roomId] == nil {
		cm.connections[roomId] = make(map[string]*websocket.Conn)
	}
	cm.connections[roomId][playerId] = conn
}

func (cm *ConnectionManager) Remove(roomId, playerId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.connections[roomId] != nil {
		delete(cm.connections[roomId], playerId)
		if len(cm.connections[roomId]) == 0 {
			delete(cm.connections, roomId)
		}
	}
}

func (cm *ConnectionManager) Broadcast(roomId string, message []byte) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if players, ok := cm.connections[roomId]; ok {
		for _, conn := range players {
			conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
