// Package connmanager registra as conexões WebSocket ativas por sala/jogador e
// publica mensagens para elas. Fica fora do motor puro (usa gorilla/websocket).
package connmanager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/gorilla/websocket"
)

// connEntry guarda uma conexão e o mutex que serializa suas escritas. O
// gorilla proíbe mais de um escritor concorrente na mesma conexão; o mutex
// garante que as escritas vindas de Broadcast/BroadcastSelective não concorram.
type connEntry struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type ConnectionManager struct {
	mu          sync.RWMutex
	connections map[string]map[string]*connEntry // roomId -> playerId -> conn
	cancels     map[string]context.CancelFunc    // roomId -> cancel
	subscribed  map[string]bool
}

var ConnManager = &ConnectionManager{
	connections: make(map[string]map[string]*connEntry),
	cancels:     make(map[string]context.CancelFunc),
	subscribed:  make(map[string]bool),
}

// Add adiciona uma nova conexão WebSocket de um jogador no mapa de conexões.
func (cm *ConnectionManager) Add(roomId, playerId string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if cm.connections[roomId] == nil {
		cm.connections[roomId] = make(map[string]*connEntry)
	}
	cm.connections[roomId][playerId] = &connEntry{conn: conn}

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

// Disconnect remove a conexão do jogador do mapa e fecha o WebSocket.
func (cm *ConnectionManager) Disconnect(roomId, playerId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.connections[roomId] != nil {
		if entry, exists := cm.connections[roomId][playerId]; exists {
			entry.conn.Close()
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

// SetRoomCancel armazena a função cancel para a sala.
func (cm *ConnectionManager) SetRoomCancel(roomId string, cancel context.CancelFunc) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.cancels[roomId] = cancel
}

// CancelRoom cancela a sala e remove todas as conexões associadas.
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

// snapshotRoom copia as conexões da sala sob RLock e devolve a cópia, liberando
// o lock global antes de qualquer escrita (evita travar a sala com um cliente lento).
func (cm *ConnectionManager) snapshotRoom(roomId string) map[string]*connEntry {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	players := cm.connections[roomId]
	snapshot := make(map[string]*connEntry, len(players))
	for playerId, entry := range players {
		snapshot[playerId] = entry
	}
	return snapshot
}

// write serializa a escrita por conexão, aplica deadline e remove a conexão em
// caso de falha sem derrubar as demais.
func (cm *ConnectionManager) write(roomId, playerId string, entry *connEntry, message []byte) {
	entry.mu.Lock()
	defer entry.mu.Unlock()

	_ = entry.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := entry.conn.WriteMessage(websocket.TextMessage, message); err != nil {
		entry.conn.Close()
		cm.removeConn(roomId, playerId)
	}
}

// removeConn remove a conexão do mapa e limpa a sala caso tenha ficado vazia.
func (cm *ConnectionManager) removeConn(roomId, playerId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.connections[roomId] == nil {
		return
	}
	delete(cm.connections[roomId], playerId)

	if len(cm.connections[roomId]) == 0 {
		if cancel, ok := cm.cancels[roomId]; ok {
			cancel()
			delete(cm.cancels, roomId)
		}
		delete(cm.connections, roomId)
		delete(cm.subscribed, roomId)
	}
}

// Ping envia um ping para uma conexão específica serializando a escrita pelo
// mesmo mutex de write (o gorilla proíbe escritores concorrentes na conexão).
func (cm *ConnectionManager) Ping(roomId, playerId string) error {
	cm.mu.RLock()
	players := cm.connections[roomId]
	var entry *connEntry
	if players != nil {
		entry = players[playerId]
	}
	cm.mu.RUnlock()

	if entry == nil {
		return fmt.Errorf("conexão não encontrada")
	}

	entry.mu.Lock()
	defer entry.mu.Unlock()

	_ = entry.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return entry.conn.WriteMessage(websocket.PingMessage, nil)
}

// Broadcast envia uma mensagem para todos os jogadores conectados em uma sala.
func (cm *ConnectionManager) Broadcast(roomId string, message []byte) {
	for playerId, entry := range cm.snapshotRoom(roomId) {
		cm.write(roomId, playerId, entry, message)
	}
}

// BroadcastSelective envia uma mensagem apenas para jogadores específicos de uma sala.
func (cm *ConnectionManager) BroadcastSelective(roomId string, message []byte, playerIds []string) {
	targetPlayers := make(map[string]bool, len(playerIds))
	for _, playerId := range playerIds {
		targetPlayers[playerId] = true
	}

	for playerId, entry := range cm.snapshotRoom(roomId) {
		if targetPlayers[playerId] {
			cm.write(roomId, playerId, entry, message)
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
