package players

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
)

// GetCardByIndex retorna a carta do jogador pelo Card.Index fornecido
func (p *Player) GetCardByIndex(index int) (*structs.Card, error) {
	for i := range p.Cards {
		if p.Cards[i].Index == index {
			return &p.Cards[i], nil
		}
	}
	return nil, fmt.Errorf("carta não encontrada: índice %d do jogador %s", index, p.Id)
}

// GetAliveCardsIndexes retorna os índices das cartas vivas do jogador
func (p *Player) GetAliveCardsIndexes() []int {
	aliveIndexes := []int{}
	for i := range p.Cards {
		if !p.Cards[i].Dead {
			aliveIndexes = append(aliveIndexes, p.Cards[i].Index)
		}
	}
	return aliveIndexes
}

// GetDeadCardsIndexes retorna os índices das cartas mortas do jogador
func (p *Player) GetDeadCardsIndexes() []int {
	deadIndexes := []int{}
	for i := range p.Cards {
		if p.Cards[i].Dead {
			deadIndexes = append(deadIndexes, p.Cards[i].Index)
		}
	}
	return deadIndexes
}

// HasDeadCard indica se o jogador possui ao menos uma carta morta na mão
func (p *Player) HasDeadCard() bool {
	return len(p.GetDeadCardsIndexes()) > 0
}

// GetFirstAliveCardIndexByName retorna o índice da primeira carta viva com o nome informado
func (p *Player) GetFirstAliveCardIndexByName(name structs.CardName) (int, bool) {
	for i := range p.Cards {
		if !p.Cards[i].Dead && p.Cards[i].Name == name {
			return p.Cards[i].Index, true
		}
	}
	return 0, false
}

// HasAliveCardName indica se o jogador possui uma carta viva com o nome informado
func (p *Player) HasAliveCardName(name structs.CardName) bool {
	_, ok := p.GetFirstAliveCardIndexByName(name)
	return ok
}

// KillCard marca a carta como morta pelo Card.Index
func (p *Player) KillCard(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}
	card.Dead = true
	card.Protected = false
	return nil
}

// ReviveCard marca a carta como viva pelo Card.Index, escondendo novamente a carta
func (p *Player) ReviveCard(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}

	card.Dead = false
	card.RevealedToAll = false
	card.Protected = false
	return nil
}

// RevealCardToAll marca a carta como revelada para todos pelo Card.Index
func (p *Player) RevealCardToAll(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}
	card.RevealedToAll = true
	return nil
}

// ProtectCard marca a carta como protegida pelo Card.Index
func (p *Player) ProtectCard(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}
	card.Protected = true
	return nil
}

// GetProtectedCardsIndexes retorna os índices das cartas protegidas do jogador
func (p *Player) GetProtectedCardsIndexes() []int {
	protectedIndexes := []int{}
	for i := range p.Cards {
		if p.Cards[i].Protected {
			protectedIndexes = append(protectedIndexes, p.Cards[i].Index)
		}
	}
	return protectedIndexes
}

// UnprotectCard remove a proteção da carta pelo Card.Index
func (p *Player) UnprotectCard(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}
	card.Protected = false
	return nil
}
