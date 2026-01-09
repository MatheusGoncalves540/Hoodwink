package roomStructs

import (
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
)

type Room struct {
	ID             string            `json:"id"`
	Rules          Rules             `json:"rules"`
	TimeoutType    string            `json:"timeoutType"`
	Name           string            `json:"name"`
	Password       string            `json:"password" validate:"max=24"`
	MaxPlayers     int               `json:"maxPlayers"`
	CustomMatch    bool              `json:"customMatch"`
	Turn           int               `json:"turn"`
	Tax            int               `json:"tax"`
	Players        map[string]Player `json:"players"`
	DeadDeck       []string          `json:"deadDeck"`
	CurrentPlayer  string            `json:"currentPlayer"`
	GameEvent      *GameEvent        `json:"gameEvent"`
	PendingEffects []Effect          `json:"pendingEffects"`
	PendingPlayer  string            `json:"pendingPlayer"`
	GameOver       bool              `json:"gameOver"`
	StartTime      time.Time         `json:"startTime"`
	Created        time.Time         `json:"created"`
}

// GetPlayer retorna o ponteiro do jogador pela playerId e um bool indicando se existe
func (r *Room) GetPlayer(playerId string) (*Player, error) {
	player, exists := r.Players[playerId]
	if !exists {
		return nil, fmt.Errorf("jogador alvo não encontrado: %s", playerId)
	}
	return &player, nil
}

// GetCardRules retorna as regras da carta específica para a sala
func (r *Room) GetCardRules(registryRules *rules.Registry, card string) (*rules.CardRules, error) {
	gameRules, err := registryRules.Get(string(r.Rules))
	if err != nil {
		return nil, err
	}
	cardRules := rules.CardRules(gameRules.Cards[card])
	return &cardRules, nil
}

// GetTimeoutDuration retorna a duração do timeout específico
func (r *Room) GetTimeoutDuration(registryRules *rules.Registry, timeoutField string) (time.Duration, error) {
	// Obtém as regras do jogo para a sala
	gameRules, err := registryRules.Get(string(r.Rules))
	if err != nil {
		return 0, err
	}

	// Obtém a configuração de timeout para o tipo da sala
	timeoutConfig, exists := gameRules.Timeouts[r.TimeoutType]
	if !exists {
		return 0, fmt.Errorf("tipo de timeout '%s' não encontrado", r.TimeoutType)
	}

	// Retorna o valor do campo via função helper
	value, err := timeoutConfig.Get(timeoutField)
	if err != nil {
		return 0, fmt.Errorf("erro ao obter timeoutField '%s': %w", timeoutField, err)
	}
	return time.Duration(value), nil
}
