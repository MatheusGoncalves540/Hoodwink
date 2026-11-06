package structs

import (
	"context"
	"fmt"
	"sync"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
)

type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]map[string]*websocket.Conn // roomId -> playerId -> conn
	cancels     map[string]context.CancelFunc         // roomId -> cancel
	subscribed  map[string]bool
}

var ConnManager = &ConnectionManager{
	connections: make(map[string]map[string]*websocket.Conn),
	cancels:     make(map[string]context.CancelFunc),
	subscribed:  make(map[string]bool),
}

// Adiciona uma nova conexão WebSocket de um jogador no mapa de conexões
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
		for rId, IsSubscribed := range cm.subscribed {
			utils.LogDebug(fmt.Sprintf("Room: %s, Subscribed: %t", rId, IsSubscribed))
		}
		utils.LogDebug("------------------------------------------------")
	}
}

// Remove a conexão do jogador do mapa e fecha o WebSocket
func (cm *ConnectionManager) Disconnect(roomId, playerId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.connections[roomId] != nil {
		if conn, exists := cm.connections[roomId][playerId]; exists {
			conn.Close()
			delete(cm.connections[roomId], playerId)
		}

		if len(cm.connections[roomId]) == 0 {
			if cancel, ok := cm.cancels[roomId]; ok {
				cancel()
				delete(cm.cancels, roomId)
			}
			delete(cm.connections, roomId)
			delete(cm.subscribed, roomId)
		}
	}

	// Debug log das conexões
	if utils.IsDebugMode() {
		utils.LogDebug("------------ Debug ConnectionManager ------------")
		for rId, players := range cm.connections {
			utils.LogDebug(fmt.Sprintf("Room: %s", rId))
			for pId := range players {
				utils.LogDebug(fmt.Sprintf("  Player: %s", pId))
			}
		}
		for rId, IsSubscribed := range cm.subscribed {
			utils.LogDebug(fmt.Sprintf("Room: %s, Subscribed: %t", rId, IsSubscribed))
		}
		utils.LogDebug("------------------------------------------------")
	}
}

// Armazena a função cancel para a sala
func (cm *ConnectionManager) SetRoomCancel(roomId string, cancel context.CancelFunc) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cancels[roomId] = cancel
}

// Cancela a sala e remove todas as conexões associadas
func (cm *ConnectionManager) CancelRoom(roomId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cancel, ok := cm.cancels[roomId]; ok {
		cancel()
		delete(cm.cancels, roomId)
	}

	delete(cm.connections, roomId)
	delete(cm.subscribed, roomId)

	utils.LogDebug("Sala " + roomId + " encerrada e cancelada")
}

// Envia uma mensagem para todos os jogadores conectados em uma sala
func (cm *ConnectionManager) Broadcast(roomId string, message []byte) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if players, ok := cm.connections[roomId]; ok {
		for _, conn := range players {
			conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// Envia uma mensagem apenas para jogadores específicos de uma sala
func (cm *ConnectionManager) BroadcastSelective(roomId string, message []byte, playerIds []string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Cria um map para busca eficiente dos playerIds
	targetPlayers := make(map[string]bool, len(playerIds))
	for _, playerId := range playerIds {
		targetPlayers[playerId] = true
	}

	if players, ok := cm.connections[roomId]; ok {
		for playerId, conn := range players {
			// Envia apenas se o playerId estiver na lista de alvos
			if targetPlayers[playerId] {
				conn.WriteMessage(websocket.TextMessage, message)
			}
		}
	}
}

func (cm *ConnectionManager) IsSubscribed(roomId string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.subscribed[roomId]
}

func (cm *ConnectionManager) MarkSubscribed(roomId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.subscribed[roomId] = true
}

func (cm *ConnectionManager) ClearSubscribed(roomId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.subscribed, roomId)
}
