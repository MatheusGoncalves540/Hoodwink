package rooms

import (
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// GetCardRules retorna as regras da carta específica para a sala
func (r *Room) GetCardRules(registryRules *rules.Registry, card string) (*rules.CardRules, error) {
	gameRules, err := registryRules.Get(string(r.Rules))
	if err != nil {
		return nil, err
	}
	cardRules := rules.CardRules(gameRules.Cards[card])
	return &cardRules, nil
}

// GetGeneralRules retorna as regras gerais do jogo para a sala
func (r *Room) GetGeneralRules(registryRules *rules.Registry) (*rules.Generals, error) {
	// Obtém as regras do jogo para a sala
	gameRules, err := registryRules.Get(string(r.Rules))
	if err != nil {
		return nil, err
	}
	return &gameRules.Generals, nil
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

// VerifyPlayerHasEnoughCoins verifica se jogador tem moedas suficientes para interagir com a carta
func (r *Room) VerifyPlayerHasEnoughCoins(player *players.Player, registryRules *rules.Registry, card structs.TypePlayerPlays, discount int) error {
	cardPrice, err := r.GetCardValue(registryRules, card, discount)
	if err != nil {
		return err
	}

	if player.Coins < cardPrice {
		return fmt.Errorf("jogador %s não tem moedas suficientes para interagir com a carta %s (tem %d, precisa de %d)", player.Id, card, player.Coins, cardPrice)
	}

	return nil
}

// MarkCardValueAsDoubled marca como dobrado o preço da carta específica
func (r *Room) MarkCardValueAsDoubled(registryRules *rules.Registry, card structs.TypePlayerPlays) {
	timesDoubledCardConf, exist := r.DoubledCardValues[card]
	if !exist {
		return
	}

	cardRules, err := r.GetCardRules(registryRules, string(card))
	if err != nil {
		utils.LogError(err)
		return
	}

	if timesDoubledCardConf.TimesValueDoubled < *cardRules.MaxDoubled {
		// ainda pode dobrar o preço
		timesDoubledCardConf.TimesValueDoubled++
		timesDoubledCardConf.RoundsUntilDecrease = *cardRules.RoundsUntilDecrease
	} else {
		// já atingiu o máximo de vezes que pode ser dobrado, apenas reseta o contador de rounds
		timesDoubledCardConf.RoundsUntilDecrease = *cardRules.RoundsUntilDecrease
	}

	timesDoubledCardConf.UsedThisTurn = true

	r.DoubledCardValues[card] = timesDoubledCardConf
}

// DecreaseDoubledCardValuesRounds decrementa o contador de rounds para redução do preço dobrado das cartas e abaixa o preço se o contador estiver em 0
func (r *Room) DecreaseDoubledCardValuesRounds() {
	for cardType, timesDoubledCardConf := range r.DoubledCardValues {
		// decrementa o contador de rounds
		if !timesDoubledCardConf.UsedThisTurn {
			if timesDoubledCardConf.RoundsUntilDecrease > 0 {
				timesDoubledCardConf.RoundsUntilDecrease--
				r.DoubledCardValues[cardType] = timesDoubledCardConf
			} else {
				// diminui o preço dobrado em 1, se possível
				if timesDoubledCardConf.TimesValueDoubled > 0 {
					timesDoubledCardConf.TimesValueDoubled--
					r.DoubledCardValues[cardType] = timesDoubledCardConf
				}
			}
		} else {
			// reseta o flag de uso nesta rodada
			timesDoubledCardConf.UsedThisTurn = false
			r.DoubledCardValues[cardType] = timesDoubledCardConf
		}
	}
}

// GetCardValue retorna o value atual da carta, considerando se está dobrado ou não, taxas e limites.
// Taxas são aplicadas depois dos valores dobrados (ambos apenas se aplicáveis)
func (r *Room) GetCardValue(registryRules *rules.Registry, card structs.TypePlayerPlays, discount int) (int, error) {
	cardRules, err := r.GetCardRules(registryRules, string(card))
	if err != nil {
		return 0, err
	}

	value := *cardRules.Value

	// se a carta funcionar com preço dobrado, aplica o valor dobrado
	timesDoubledCardConf, exist := r.DoubledCardValues[card]
	if exist {
		value = value << timesDoubledCardConf.TimesValueDoubled
	}

	// se a carta for afetada por taxas, aplica a taxa
	if cardRules.AffectedByTaxes != nil && *cardRules.AffectedByTaxes {
		if value+r.Tax > *cardRules.TaxMaximum {
			// se a taxa ultrapassar os limites, aplica os limites
			value = *cardRules.TaxMaximum
		} else if value+r.Tax < *cardRules.TaxMinimum {
			// se a taxa for menor que o mínimo, aplica o mínimo
			value = *cardRules.TaxMinimum
		} else {
			// caso contrário, aplica a taxa normalmente
			value += r.Tax
		}
	}

	// Aplica o desconto, se houver
	value -= discount
	if value < 0 {
		value = 0
	}

	return value, nil
}

// ChangeCard troca a carta do index por uma nova carta (marca a carta como viva)
func (r *Room) ChangeCard(player *players.Player, index int) error {
	card, err := player.GetCardByIndex(index)
	if err != nil {
		return err
	}

	// troca a carta pelo topo do deck
	card.Name = r.Deck[0]
	// remove a carta do topo do deck
	r.Deck = r.Deck[1:]

	// carta sempre vem como viva
	card.Dead = false
	return nil
}
