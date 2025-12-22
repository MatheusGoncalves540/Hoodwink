package wsRoom

import (
	"fmt"
	"sync"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
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

	// Debug log das conexões
	if utils.IsDebugMode() {
		utils.LogDebug("------------ Debug ConnectionManager ------------")
		for rId, players := range cm.connections {
			utils.LogDebug(fmt.Sprintf("Room: %s", rId))
			for pId := range players {
				utils.LogDebug(fmt.Sprintf("  Player: %s", pId))
			}
		}
		utils.LogDebug("------------------------------------------------")
	}
}

func (cm *ConnectionManager) Disconnect(roomId, playerId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.connections[roomId] != nil {
		if conn, exists := cm.connections[roomId][playerId]; exists {
			conn.Close()
			delete(cm.connections[roomId], playerId)
			if len(cm.connections[roomId]) == 0 {
				delete(cm.connections, roomId)
			}
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

func (cm *ConnectionManager) GetRoomPlayers(roomId string) map[string]*websocket.Conn {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if players, ok := cm.connections[roomId]; ok {
		// Retorna uma cópia para evitar modificações externas
		copy := make(map[string]*websocket.Conn)
		for k, v := range players {
			copy[k] = v
		}
		return copy
	}
	return nil
}

func (cm *ConnectionManager) RoomExists(roomId string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	_, exists := cm.connections[roomId]
	return exists
}
