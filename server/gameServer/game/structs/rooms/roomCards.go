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
	cardRules, ok := gameRules.Cards[card]
	if !ok {
		return nil, fmt.Errorf("regras da carta '%s' não encontradas no modo %s", card, r.Rules)
	}
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

	if cardRules.MaxDoubled != nil && timesDoubledCardConf.TimesValueDoubled < *cardRules.MaxDoubled {
		// ainda pode dobrar o preço
		timesDoubledCardConf.TimesValueDoubled++
	} else {
		// já atingiu o máximo de vezes que pode ser dobrado
		timesDoubledCardConf.TimesValueDoubled = maxDoubledValue(cardRules)
	}

	if cardRules.RoundsUntilDecrease != nil {
		timesDoubledCardConf.RoundsUntilDecrease = *cardRules.RoundsUntilDecrease
	}

	timesDoubledCardConf.UsedThisTurn = true

	r.DoubledCardValues[card] = timesDoubledCardConf
}

// DecreaseDoubledCardValuesRounds decrementa o contador de turnos sem uso das cartas
// com valor dobrado. A cada N turnos sem uso (roundsUntilDecrease), o preço cai pela metade.
func (r *Room) DecreaseDoubledCardValuesRounds(registryRules *rules.Registry) {
	for cardType, timesDoubledCardConf := range r.DoubledCardValues {
		if timesDoubledCardConf.UsedThisTurn {
			// reseta o flag de uso nesta rodada
			timesDoubledCardConf.UsedThisTurn = false
			r.DoubledCardValues[cardType] = timesDoubledCardConf
			continue
		}

		decreaseInterval := 3
		if cardRules, err := r.GetCardRules(registryRules, string(cardType)); err == nil && cardRules.RoundsUntilDecrease != nil {
			decreaseInterval = *cardRules.RoundsUntilDecrease
		}

		if timesDoubledCardConf.RoundsUntilDecrease > 0 {
			timesDoubledCardConf.RoundsUntilDecrease--
		} else {
			timesDoubledCardConf.RoundsUntilDecrease = decreaseInterval
		}

		// Atingiu o intervalo sem uso: reduz o valor dobrado pela metade
		if timesDoubledCardConf.RoundsUntilDecrease == 0 && timesDoubledCardConf.TimesValueDoubled > 0 {
			timesDoubledCardConf.TimesValueDoubled--
			timesDoubledCardConf.RoundsUntilDecrease = decreaseInterval
		}

		r.DoubledCardValues[cardType] = timesDoubledCardConf
	}
}

// maxDoubledValue retorna o número máximo de vezes que a carta pode ter o valor dobrado
func maxDoubledValue(cardRules *rules.CardRules) int {
	if cardRules.MaxDoubled == nil {
		return 0
	}
	return *cardRules.MaxDoubled
}

// GetCardValue retorna o value atual da carta, considerando se está dobrado ou não, taxas e limites.
// Taxas são aplicadas depois dos valores dobrados (ambos apenas se aplicáveis)
func (r *Room) GetCardValue(registryRules *rules.Registry, card structs.TypePlayerPlays, discount int) (int, error) {
	cardRules, err := r.GetCardRules(registryRules, string(card))
	if err != nil {
		return 0, err
	}
	if cardRules.Value == nil {
		return 0, fmt.Errorf("carta %s não possui value configurado", card)
	}

	value := *cardRules.Value

	// se a carta funcionar com preço dobrado, aplica o valor dobrado
	timesDoubledCardConf, exist := r.DoubledCardValues[card]
	if exist {
		value = value << timesDoubledCardConf.TimesValueDoubled
	}

	// se a carta for afetada por taxas, aplica a taxa
	if cardRules.AffectedByTaxes != nil && *cardRules.AffectedByTaxes {
		value += r.Tax

		if cardRules.TaxMaximum != nil && value > *cardRules.TaxMaximum {
			value = *cardRules.TaxMaximum
		}
		if cardRules.TaxMinimum != nil && value < *cardRules.TaxMinimum {
			value = *cardRules.TaxMinimum
		}
	}

	// Aplica o desconto, se houver
	value -= discount
	if value < 0 {
		value = 0
	}

	return value, nil
}

// GetGuardianValue retorna o custo da Guardiã de acordo com o tipo de proteção.
// - "sight": bloqueio do Clarividente (valor reduzido e limitado pelo mínimo da carta)
// - "future": proteção contra um ataque futuro (acréscimo e mínimo próprios)
// - demais: defesa imediata de uma carta
func (r *Room) GetGuardianValue(registryRules *rules.Registry, protectedFrom string) (int, error) {
	cardRules, err := r.GetCardRules(registryRules, string(structs.PlayGuardianCard))
	if err != nil {
		return 0, err
	}
	if cardRules.Value == nil {
		return 0, fmt.Errorf("carta GUARDIAN não possui value configurado")
	}

	value := *cardRules.Value
	if cardRules.AffectedByTaxes != nil && *cardRules.AffectedByTaxes {
		value += r.Tax
	}

	switch protectedFrom {
	case "sight":
		if cardRules.ClairvoyantBlockDiscount != nil {
			value -= *cardRules.ClairvoyantBlockDiscount
		}
	case "future":
		if cardRules.FutureProtectionSurcharge != nil {
			value += *cardRules.FutureProtectionSurcharge
		}
	}

	if protectedFrom == "future" && cardRules.FutureProtectionTaxMinimum != nil {
		if value < *cardRules.FutureProtectionTaxMinimum {
			value = *cardRules.FutureProtectionTaxMinimum
		}
	} else if cardRules.TaxMinimum != nil && value < *cardRules.TaxMinimum {
		value = *cardRules.TaxMinimum
	}

	if cardRules.TaxMaximum != nil && value > *cardRules.TaxMaximum {
		value = *cardRules.TaxMaximum
	}

	if value < 0 {
		value = 0
	}
	return value, nil
}

// GetBasicTakeCoinsAmount retorna quantas moedas a ação básica de coletar concede (1 + taxas).
func (r *Room) GetBasicTakeCoinsAmount(registryRules *rules.Registry) (int, error) {
	generalRules, err := r.GetGeneralRules(registryRules)
	if err != nil {
		return 0, err
	}
	base := 1
	if generalRules.BasicTakeCoins != nil {
		base = *generalRules.BasicTakeCoins
	}

	amount := base + r.Tax
	if amount < 0 {
		amount = 0
	}
	return amount, nil
}

// GetBasicKillCardCost retorna o custo da ação básica de matar uma carta.
// A metade da taxa só é aplicada quando ela é positiva; com taxa zero ou negativa
// a taxa entra inteira. O custo nunca fica abaixo do mínimo configurado (3).
func (r *Room) GetBasicKillCardCost(registryRules *rules.Registry) (int, error) {
	generalRules, err := r.GetGeneralRules(registryRules)
	if err != nil {
		return 0, err
	}
	base := 6
	if generalRules.BasicKillCardCost != nil {
		base = *generalRules.BasicKillCardCost
	}
	minimum := 3
	if generalRules.BasicKillCardMinimum != nil {
		minimum = *generalRules.BasicKillCardMinimum
	}

	cost := base
	if r.Tax > 0 {
		cost += utils.FloorDiv(r.Tax, 2)
	} else {
		cost += r.Tax
	}
	if cost < minimum {
		cost = minimum
	}
	return cost, nil
}

// GetBasicSwapCost retorna o custo da ação básica de trocar cartas com o baralho.
// Trocar uma carta custa metade das moedas arredondada para cima; trocar as duas custa todas as moedas.
func (r *Room) GetBasicSwapCost(coins int, twoCards bool) int {
	if twoCards {
		return coins
	}
	return utils.CeilDiv(coins, 2)
}

// EnforceInvestorMaxInvestments descarta investimentos excedentes de todos os jogadores
// quando o limite máximo de investimentos (afetado pelas taxas) diminui.
func (r *Room) EnforceInvestorMaxInvestments(registryRules *rules.Registry) {
	maxInvestments, err := r.GetCardValue(registryRules, structs.PlayInvestorCard, 0)
	if err != nil {
		utils.LogError(err)
		return
	}

	for _, player := range r.Players {
		player.TrimInvestments(maxInvestments)
	}
}

// ChangeCard troca a carta do index por uma nova carta (marca a carta como viva)
func (r *Room) ChangeCard(player *players.Player, index int) error {
	card, err := player.GetCardByIndex(index)
	if err != nil {
		return err
	}

	if len(r.Deck) == 0 {
		return fmt.Errorf("baralho vazio: não é possível trocar a carta %d do jogador %s", index, player.Id)
	}

	// troca a carta pelo topo do deck
	card.Name = r.Deck[0]
	// remove a carta do topo do deck
	r.Deck = r.Deck[1:]

	// carta sempre vem como viva e virada para baixo
	card.Dead = false
	card.RevealedToAll = false
	card.Protected = false
	return nil
}

// ChangeAllCards troca todas as cartas de todos os jogadores por novas cartas
func (r *Room) ChangeAllCards() error {
	for _, player := range r.Players {
		for index := range player.Cards {
			err := r.ChangeCard(player, index)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
