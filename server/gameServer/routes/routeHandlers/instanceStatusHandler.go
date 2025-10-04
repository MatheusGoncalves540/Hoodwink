package routeHandlers

import (
	"context"
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

type InstanceStatus struct {
	InstanceID     string     `json:"instanceId"`
	IsAlive        bool       `json:"isAlive"`
	ManagedPlayers []string   `json:"managedPlayers"`
	LastSeen       *time.Time `json:"lastSeen,omitempty"`
}

type StatusResponse struct {
	CurrentInstance string           `json:"currentInstance"`
	TotalInstances  int              `json:"totalInstances"`
	Instances       []InstanceStatus `json:"instances"`
	OrphanedPlayers []string         `json:"orphanedPlayers"`
}

// GetInstancesStatus retorna o status de todas as instâncias
func (h *Handler) GetInstancesStatus(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		status := StatusResponse{
			CurrentInstance: utils.GetInstanceID(),
			Instances:       []InstanceStatus{},
			OrphanedPlayers: []string{},
		}

		// Busca todas as instâncias vivas
		aliveInstances := services.GetAliveInstances(ctx, rdb)
		status.TotalInstances = len(aliveInstances)

		// Para cada instância, coleta informações
		for _, instanceID := range aliveInstances {
			instanceStatus := InstanceStatus{
				InstanceID:     instanceID,
				IsAlive:        true,
				ManagedPlayers: getManagedPlayersByInstance(ctx, rdb, instanceID),
			}
			status.Instances = append(status.Instances, instanceStatus)
		}

		// Busca players órfãos (instâncias mortas)
		status.OrphanedPlayers = getOrphanedPlayers(ctx, rdb, aliveInstances)

		utils.SendJSON(w, http.StatusOK, utils.APIResponse{
			Message: "Status das instâncias",
			Data:    status,
		})
	}
}

// getManagedPlayersByInstance retorna players gerenciados por uma instância específica
func getManagedPlayersByInstance(ctx context.Context, rdb *redis.Client, instanceID string) []string {
	var players []string

	keys, err := rdb.Keys(ctx, "player:*:room").Result()
	if err != nil {
		return players
	}

	for _, key := range keys {
		value, err := rdb.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		// Verifica se o valor termina com a instanceID
		suffix := ":" + instanceID
		if len(value) > len(suffix) && value[len(value)-len(suffix):] == suffix {
			playerId := key[7 : len(key)-5] // remove "player:" e ":room"
			players = append(players, playerId)
		}
	}

	return players
}

// getOrphanedPlayers retorna players de instâncias mortas
func getOrphanedPlayers(ctx context.Context, rdb *redis.Client, aliveInstances []string) []string {
	var orphaned []string

	keys, err := rdb.Keys(ctx, "player:*:room").Result()
	if err != nil {
		return orphaned
	}

	aliveMap := make(map[string]bool)
	for _, instanceID := range aliveInstances {
		aliveMap[instanceID] = true
	}

	for _, key := range keys {
		value, err := rdb.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		// Parse da instanceID
		lastColon := -1
		for i := len(value) - 1; i >= 0; i-- {
			if value[i] == ':' {
				lastColon = i
				break
			}
		}

		if lastColon == -1 {
			continue
		}

		instanceID := value[lastColon+1:]
		if !aliveMap[instanceID] {
			playerId := key[7 : len(key)-5] // remove "player:" e ":room"
			orphaned = append(orphaned, playerId)
		}
	}

	return orphaned
}

// CleanupOrphanedPlayers remove registros de players órfãos
func (h *Handler) CleanupOrphanedPlayers(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := utils.CleanupPlayerRegistrations(ctx, rdb)
		if err != nil {
			utils.SendError(w, "Erro ao limpar registros órfãos", http.StatusInternalServerError)
			return
		}

		utils.SendJSON(w, http.StatusOK, utils.APIResponse{
			Message: "Limpeza de registros órfãos concluída",
		})
	}
}
