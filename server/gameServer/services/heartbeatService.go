package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

var HeartbeatInterval = time.Duration(utils.MustEnvInt("HEARTBEAT_INTERVAL", 10)) * time.Second
var HeartbeatTTL = time.Duration(utils.MustEnvInt("HEARTBEAT_TTL", 30)) * time.Second

type HeartbeatService struct {
	rdb        *redis.Client
	instanceID string
	ctx        context.Context
	cancel     context.CancelFunc
	ticker     *time.Ticker
}

// NewHeartbeatService cria uma nova instância do serviço de heartbeat
func NewHeartbeatService(rdb *redis.Client) *HeartbeatService {
	ctx, cancel := context.WithCancel(context.Background())
	return &HeartbeatService{
		rdb:        rdb,
		instanceID: config.InstanceID,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start inicia o heartbeat da instância
func (h *HeartbeatService) Start() {
	h.ticker = time.NewTicker(HeartbeatInterval)

	// Primeira execução imediata
	h.sendHeartbeat()

	go func() {
		for {
			select {
			case <-h.ticker.C:
				h.sendHeartbeat()
			case <-h.ctx.Done():
				h.ticker.Stop()
				return
			}
		}
	}()

	utils.LogDebug(fmt.Sprintf("Heartbeat iniciado para instância %s", h.instanceID))
}

// Stop para o heartbeat
func (h *HeartbeatService) Stop() {
	if h.cancel != nil {
		h.cancel()
	}
	if h.ticker != nil {
		h.ticker.Stop()
	}

	// Remove o heartbeat da instância do Redis
	h.rdb.Del(h.ctx, fmt.Sprintf("instance:%s:alive", h.instanceID))
	utils.LogDebug(fmt.Sprintf("Heartbeat parado para instância %s", h.instanceID))
}

// sendHeartbeat envia o heartbeat e renova TTL dos players gerenciados
func (h *HeartbeatService) sendHeartbeat() {
	// Renova heartbeat da instância
	heartbeatKey := fmt.Sprintf("instance:%s:alive", h.instanceID)
	err := h.rdb.Set(h.ctx, heartbeatKey, "1", HeartbeatTTL).Err()
	if err != nil {
		utils.LogDebug(fmt.Sprintf("Erro ao enviar heartbeat: %v", err))
		return
	}

	// Busca todos os players gerenciados por esta instância
	players := h.getManagedPlayers()

	playersRenewed := 0
	for _, playerId := range players {
		// Monta a chave do player
		playerKey := fmt.Sprintf("player:%s:room", playerId)

		// Busca o valor da chave (formato: roomId:instanceId)
		value, err := h.rdb.Get(h.ctx, playerKey).Result()
		if err != nil || value == "" {
			continue // Não encontrado
		}

		// Separa roomId e instanceId
		parts := strings.SplitN(value, ":", 2)
		if len(parts) < 2 {
			continue // Formato inesperado
		}
		roomId := parts[0]

		// Verifica se a sala existe
		roomKey := fmt.Sprintf("room:%s", roomId)
		exists, err := h.rdb.Exists(h.ctx, roomKey).Result()
		if err != nil || exists == 0 {
			continue // Sala não existe
		}

		// Renova TTL do player
		h.rdb.Expire(h.ctx, playerKey, HeartbeatTTL)
		playersRenewed++
	}

	if playersRenewed > 0 {
		utils.LogDebug(fmt.Sprintf(
			"Heartbeat enviado - Instância: %s, Players renovados: %d",
			h.instanceID, playersRenewed,
		))
	}
}

// getManagedPlayers retorna a lista de players gerenciados por esta instância
func (h *HeartbeatService) getManagedPlayers() []string {
	var players []string

	// Busca todas as chaves de player:*:room
	keys, err := h.rdb.Keys(h.ctx, "player:*:room").Result()
	if err != nil {
		return players
	}

	// Para cada chave, verifica se o valor contém nossa instanceID
	for _, key := range keys {
		value, err := h.rdb.Get(h.ctx, key).Result()
		if err != nil {
			continue
		}

		// Verifica se o valor termina com nossa instanceID (formato: roomId:instanceId)
		suffix := ":" + h.instanceID
		if len(value) > len(suffix) && value[len(value)-len(suffix):] == suffix {
			// Extrai playerId da chave "player:{playerId}:room"
			playerId := key[7 : len(key)-5] // remove "player:" e ":room"
			players = append(players, playerId)
		}
	}

	return players
}

// IsInstanceAlive verifica se uma instância específica está viva
func IsInstanceAlive(ctx context.Context, rdb *redis.Client, instanceID string) bool {
	key := fmt.Sprintf("instance:%s:alive", instanceID)
	_, err := rdb.Get(ctx, key).Result()
	return err == nil
}

// GetAliveInstances retorna lista de instâncias vivas
func GetAliveInstances(ctx context.Context, rdb *redis.Client) []string {
	var instances []string

	keys, err := rdb.Keys(ctx, "instance:*:alive").Result()
	if err != nil {
		return instances
	}

	for _, key := range keys {
		// Extrai instanceID da chave "instance:{instanceID}:alive"
		instanceID := key[9 : len(key)-6] // remove "instance:" e ":alive"
		instances = append(instances, instanceID)
	}

	return instances
}
