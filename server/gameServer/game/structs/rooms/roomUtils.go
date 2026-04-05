package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
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

// RandomlyChangeTaxes aleatoriamente aumenta ou diminui as taxas
func (r *Room) RandomlyChangeTaxes(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry) error {
	generalRules, err := r.GetGeneralRules(registryRules)
	if err != nil {
		return err
	}

	randPercentage := rand.IntN(99) + 1 // gera um número entre 1 e 100

	if randPercentage <= *generalRules.RandomTaxesChanges {
		// Taxas serão alteradas

		// Decide aleatoriamente se as taxas aumentam ou diminuem
		//TODO se pá é interessante fazer um anuncer para isso, ou pelo menos logar
		if rand.IntN(2) == 0 {
			// Aumenta as taxas
			r.IncrementTax(1)
			utils.LogDebug("Taxas aumentadas para" + fmt.Sprint(r.Tax))
		} else {
			// Diminui as taxas
			r.DecrementTax(1)
			utils.LogDebug("Taxas diminuídas para" + fmt.Sprint(r.Tax))
		}
	}

	return nil
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
