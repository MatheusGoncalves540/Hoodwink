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
func (p *Player) GetCardByIndex(index int) (*Card, bool) {
	for i := range p.Cards {
		if p.Cards[i].Index == index {
			return &p.Cards[i], true
		}
	}
	return nil, false
}

// KillCard marca a carta como morta pelo Card.Index
func (p *Player) KillCard(index int) error {
	for i := range p.Cards {
		if p.Cards[i].Index == index {
			p.Cards[i].Dead = true
			return nil
		}
	}
	return fmt.Errorf("carta não encontrada com index %d", index)
}
