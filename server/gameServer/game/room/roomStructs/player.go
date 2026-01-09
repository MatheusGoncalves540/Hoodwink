package roomStructs

import "fmt"

type Player struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Cards []Card `json:"cards"`
	Coins int    `json:"coins"`
	Alive bool   `json:"alive"`
}

// GetCardByIndex retorna a carta do jogador pelo Card.Index fornecido
func (p *Player) GetCardByIndex(index int) (*Card, error) {
	for i := range p.Cards {
		if p.Cards[i].Index == index {
			return &p.Cards[i], nil
		}
	}
	return nil, fmt.Errorf("carta não encontrada: índice %d do jogador %s", index, p.Id)
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

// ProtectCard marca a carta como protegida pelo Card.Index
func (p *Player) ProtectCard(index int) error {
	card, err := p.GetCardByIndex(index)
	if err != nil {
		return err
	}
	card.Protected = true
	return nil
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

// AddCoins adiciona moedas ao jogador
func (p *Player) AddCoins(amount int) {
	p.Coins += amount
	// TODO: adicionar verificação de limite máximo de 20 moedas
}

// RemoveCoins remove moedas do jogador
func (p *Player) RemoveCoins(amount int) {
	p.Coins -= amount
}
