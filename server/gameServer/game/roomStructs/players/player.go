package players

import (
	"context"
	"fmt"
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/redis/go-redis/v9"
)

type Player struct {
	Id    string             `json:"id"`
	Name  string             `json:"name"`
	Cards []roomStructs.Card `json:"cards"`
	Coins int                `json:"coins"`
	Alive bool               `json:"alive"`
}

type PublicPlayerForUpdates struct {
	Id    string                             `json:"id"`
	Name  string                             `json:"name"`
	Cards []roomStructs.PublicCardForUpdates `json:"cards"`
	Coins int                                `json:"coins"`
	Alive bool                               `json:"alive"`
}

// GetPublicPlayerForUpdates retorna a versão pública do jogador para updates
func (p *Player) GetPublicPlayerForUpdates() PublicPlayerForUpdates {
	PublicCardInfos := make([]roomStructs.PublicCardForUpdates, len(p.Cards))
	for i := range p.Cards {
		PublicCardInfos[i] = roomStructs.PublicCardForUpdates{
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
func (p *Player) GetCardByIndex(index int) (*roomStructs.Card, error) {
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

// GetPlayerRegistrationInfo retorna informações completas do registro do player
// Retorno: roomId, instanceId, bool (registrado), erro
func (p *Player) GetPlayerRegistrationInfo(ctx context.Context, rdb *redis.Client) (string, string, bool, error) {
	value, err := rdb.Get(ctx, "player:"+p.Id+":room").Result()
	if err == redis.Nil {
		return "", "", false, nil
	}
	if err != nil {
		return "", "", false, err
	}

	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return "", "", false, nil
	}

	return parts[0], parts[1], true, nil
}
