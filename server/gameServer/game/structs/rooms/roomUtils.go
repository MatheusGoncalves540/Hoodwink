package rooms

import (
	"context"
	"encoding/json"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/redis/go-redis/v9"
)

// IncrementTax incrementa as taxas pelo valor passado
func (r *Room) IncrementTax(amount int) {
	r.Tax += amount
}

// DecrementTax decrementa as taxas pelo valor passado
func (r *Room) DecrementTax(amount int) {
	r.Tax -= amount
}

// CheckIfItsEmpty verifica se uma sala está vazia (sem jogadores conectados)
func (r *Room) CheckIfItsEmpty(ctx context.Context, rdb *redis.Client) (bool, error) {
	// Verifica se há jogadores conectados (registrados nesta sala)
	for playerId := range r.Players {
		registeredRoom, registered, err := playerRedis.GetRegisteredRoomForPlayer(ctx, rdb, playerId)
		if err != nil {
			return false, err
		}
		if registered && registeredRoom == r.ID {
			return false, nil
		}
	}

	return true, nil
}

// Clone cria uma cópia profunda da sala (usando serialização JSON)
func (r *Room) Clone() *Room {
	var clone Room
	b, _ := json.Marshal(r)
	_ = json.Unmarshal(b, &clone)
	return &clone
}
