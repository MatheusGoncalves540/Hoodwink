package players

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

type Player struct {
	Id    string         `json:"id"`
	Name  string         `json:"name"`
	Cards []structs.Card `json:"cards"`
	Coins int            `json:"coins"`
	Alive bool           `json:"alive"`
}

type PublicPlayerForUpdates struct {
	Id    string                         `json:"id"`
	Name  string                         `json:"name"`
	Cards []structs.PublicCardForUpdates `json:"cards"`
	Coins int                            `json:"coins"`
	Alive bool                           `json:"alive"`
}

// GetPublicPlayerForUpdates retorna a versão pública do jogador para updates
func (p *Player) GetPublicPlayerForUpdates() PublicPlayerForUpdates {
	PublicCardInfos := make([]structs.PublicCardForUpdates, len(p.Cards))
	for i := range p.Cards {
		PublicCardInfos[i] = structs.PublicCardForUpdates{
			Index:     p.Cards[i].Index,
			Protected: p.Cards[i].Protected,
			Dead:      p.Cards[i].Dead,
		}
	}

	return PublicPlayerForUpdates{
		Id:    p.Id,
		Name:  p.Name,
		Cards: PublicCardInfos,
		Coins: p.Coins,
		Alive: p.Alive,
	}
}

// GetCardByIndex retorna a carta do jogador pelo Card.Index fornecido
func (p *Player) GetCardByIndex(index int) (*structs.Card, error) {
	for i := range p.Cards {
		if p.Cards[i].Index == index {
			return &p.Cards[i], nil
		}
	}
	return nil, fmt.Errorf("carta não encontrada: índice %d do jogador %s", index, p.Id)
}

// GetAliveCards retorna os índices das cartas vivas do jogador
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

// AddCoins adiciona moedas ao jogador
func (p *Player) AddCoins(amount int, maxCoins int) bool {
	// breakLimit indica se o limite de moedas será ultrapassado
	breakLimit := false
	if p.Coins+amount > maxCoins {
		breakLimit = true
	}

	p.Coins += amount

	return breakLimit
}

// RemoveCoins remove moedas do jogador
func (p *Player) RemoveCoins(amount int) {
	if amount > p.Coins {
		p.Coins = 0
		return
	}
	p.Coins -= amount
}
