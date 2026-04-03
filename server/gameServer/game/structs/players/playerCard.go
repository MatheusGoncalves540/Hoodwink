package players

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
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

// KillCard marca a carta como morta pelo Card.Index
func (p *Player) KillCard(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}
	card.Dead = true
	return nil
}

// ReviveCard marca a carta como viva pelo Card.Index
func (p *Player) ReviveCard(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}

	card.Dead = false
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
