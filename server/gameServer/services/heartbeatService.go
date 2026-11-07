package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

type HeartbeatService struct {
	rdb               *redis.Client
	instanceID        string
	ctx               context.Context
	cancel            context.CancelFunc
	ticker            *time.Ticker
	HeartbeatInterval time.Duration
	HeartbeatTTL      time.Duration
}

// NewHeartbeatService cria uma nova instância do serviço de heartbeat
func NewHeartbeatService(rdb *redis.Client) *HeartbeatService {
	ctx, cancel := context.WithCancel(context.Background())
	HeartbeatInterval := time.Duration(utils.MustEnvInt("HEARTBEAT_INTERVAL", 10)) * time.Second
	HeartbeatTTL := time.Duration(utils.MustEnvInt("HEARTBEAT_TTL", 30)) * time.Second

	return &HeartbeatService{
		rdb:               rdb,
		instanceID:        utils.GetInstanceID(),
		ctx:               ctx,
		cancel:            cancel,
		HeartbeatInterval: HeartbeatInterval,
		HeartbeatTTL:      HeartbeatTTL,
	}
}

// Start inicia o heartbeat da instância
func (h *HeartbeatService) Start() {
	h.ticker = time.NewTicker(h.HeartbeatInterval)

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

	utils.LogDebug(fmt.Sprintf("❤ Heartbeat iniciado para instância %s", h.instanceID))
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

// sendHeartbeat envia o heartbeat e renova TTL dos players gerenciados.
// Usa MGET (players) e pipelines (EXISTS/EXPIRE) para não fazer 3 comandos por jogador.
func (h *HeartbeatService) sendHeartbeat() {
	// Renova heartbeat da instância
	heartbeatKey := fmt.Sprintf("instance:%s:alive", h.instanceID)
	err := h.rdb.Set(h.ctx, heartbeatKey, "1", h.HeartbeatTTL).Err()
	if err != nil {
		utils.LogError(fmt.Sprintf("Erro ao enviar heartbeat: %v", err))
		return
	}

	// Busca todos os players gerenciados por esta instância (playerId -> roomId)
	players := h.getManagedPlayers()
	if len(players) == 0 {
		return
	}

	// Uma checagem de existência por sala (compartilhada entre os jogadores dela),
	// em um único pipeline.
	existsCmds := make(map[string]*redis.IntCmd, len(players))
	pipe := h.rdb.Pipeline()
	for _, roomId := range players {
		if _, ok := existsCmds[roomId]; ok {
			continue
		}
		existsCmds[roomId] = pipe.Exists(h.ctx, "room:"+roomId)
	}
	_, _ = pipe.Exec(h.ctx)

	// Renova o TTL dos jogadores cuja sala ainda existe, também em um pipeline.
	pipe = h.rdb.Pipeline()
	playersRenewed := 0
	for playerId, roomId := range players {
		if exists, err := existsCmds[roomId].Result(); err != nil || exists == 0 {
			continue // Sala não existe
		}
		pipe.Expire(h.ctx, "player:"+playerId+":room", h.HeartbeatTTL)
		playersRenewed++
	}
	if playersRenewed > 0 {
		_, _ = pipe.Exec(h.ctx)
	}

	// utils.LogDebug(fmt.Sprintf(
	// 	"Heartbeat enviado - Instância: %s, Players renovados: %d",
	// 	h.instanceID, playersRenewed,
	// ))
}

// getManagedPlayers retorna os players gerenciados por esta instância no formato
// playerId -> roomId. Usa SCAN para achar as chaves e um único MGET para ler os valores.
func (h *HeartbeatService) getManagedPlayers() map[string]string {
	players := make(map[string]string)

	// Busca todas as chaves de player:*:room via SCAN (sem KEYS bloqueante).
	keys, err := redisFuncs.ScanKeys(h.ctx, h.rdb, "player:*:room")
	if err != nil || len(keys) == 0 {
		return players
	}

	values, err := h.rdb.MGet(h.ctx, keys...).Result()
	if err != nil {
		return players
	}

	suffix := ":" + h.instanceID
	for i, key := range keys {
		value, ok := values[i].(string)
		if !ok || len(value) <= len(suffix) || !strings.HasSuffix(value, suffix) {
			continue
		}

		// Extrai playerId da chave "player:{playerId}:room" e o roomId do valor.
		playerId := key[7 : len(key)-5]
		players[playerId] = strings.TrimSuffix(value, suffix)
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

	keys, err := redisFuncs.ScanKeys(ctx, rdb, "instance:*:alive")
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
