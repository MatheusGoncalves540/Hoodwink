package rooms

import (
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/players"
)

// IncrementTax incrementa as taxas pelo valor passado
func (r *Room) IncrementTax(amount int) {
	r.Tax += amount
}

// DecrementTax decrementa as taxas pelo valor passado
func (r *Room) DecrementTax(amount int) {
	r.Tax -= amount
}

// RandomlyChangeTaxes aleatoriamente aumenta ou diminui as taxas.
// No modo clássico a chance é 0: as taxas só mudam por efeito de cartas.
func (r *Room) RandomlyChangeTaxes(registryRules *rules.Registry) error {
	generalRules, err := r.GetGeneralRules(registryRules)
	if err != nil {
		return err
	}

	if generalRules.RandomTaxesChanges == nil || *generalRules.RandomTaxesChanges <= 0 {
		return nil
	}

	randPercentage := r.Rand().IntN(99) + 1 // gera um número entre 1 e 100

	if randPercentage <= *generalRules.RandomTaxesChanges {
		// Taxas serão alteradas

		if generalRules.MaxTax != nil && r.Tax >= *generalRules.MaxTax {
			// Se as taxas já estão no máximo, só podem diminuir
			r.DecrementTax(1)
			pure.LogDebug("Taxas diminuídas para " + fmt.Sprint(r.Tax))
		} else if generalRules.MinTax != nil && r.Tax <= *generalRules.MinTax {
			// Se as taxas já estão no mínimo, só podem aumentar
			r.IncrementTax(1)
			pure.LogDebug("Taxas aumentadas para " + fmt.Sprint(r.Tax))
		} else if r.Rand().IntN(2) == 0 {
			// Aumenta as taxas
			r.IncrementTax(1)
			pure.LogDebug("Taxas aumentadas para " + fmt.Sprint(r.Tax))
		} else {
			// Diminui as taxas
			r.DecrementTax(1)
			pure.LogDebug("Taxas diminuídas para " + fmt.Sprint(r.Tax))
		}

		r.EnforceInvestorMaxInvestments(registryRules)
	}

	return nil
}

// Clone cria uma cópia profunda da sala (usando serialização JSON)
func (r *Room) Clone() *Room {
	var clone Room
	b, _ := json.Marshal(r)
	_ = json.Unmarshal(b, &clone)
	return &clone
}

// Clone cria uma cópia profunda da sala (usando serialização JSON)
func (r *PublicRoomForUpdates) Clone() *PublicRoomForUpdates {
	var clone PublicRoomForUpdates
	b, _ := json.Marshal(r)
	_ = json.Unmarshal(b, &clone)
	return &clone
}

// AppendChatMessage adiciona uma mensagem ao chat da sala, mantendo apenas as últimas 30 mensagens para evitar crescimento ilimitado
func (r *Room) AppendChatMessage(player *players.Player, message string) error {
	r.Chat = append(r.Chat, structs.ChatMessage{
		PlayerName: player.Name,
		Timestamp:  r.Now(),
		Message:    message,
	})

	// Limita o chat para as últimas 30 mensagens
	if len(r.Chat) > 30 {
		r.Chat = r.Chat[len(r.Chat)-30:]
	}

	return nil
}
