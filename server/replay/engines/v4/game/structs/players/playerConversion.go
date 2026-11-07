package players

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs"
)

// cloneInvestments garante que o estado público nunca serialize investments como null
func cloneInvestments(investments []int) []int {
	clone := make([]int, len(investments))
	copy(clone, investments)
	return clone
}

// GetPublicPlayerForUpdates retorna a versão pública do jogador para updates
func (p *Player) GetPublicPlayerForUpdates() PublicPlayerForUpdates {
	PublicCardInfos := make([]structs.PublicCardForUpdates, len(p.Cards))
	for i := range p.Cards {
		publicCard := structs.PublicCardForUpdates{
			Index:     p.Cards[i].Index,
			Protected: p.Cards[i].Protected,
			Dead:      p.Cards[i].Dead,
		}
		if p.Cards[i].Dead || p.Cards[i].RevealedToAll {
			name := p.Cards[i].Name
			publicCard.Name = &name
		}
		PublicCardInfos[i] = publicCard
	}

	return PublicPlayerForUpdates{
		Id:                    p.Id,
		Name:                  p.Name,
		MoveTime:              p.MoveTime,
		Cards:                 PublicCardInfos,
		Coins:                 p.Coins,
		Investments:           cloneInvestments(p.Investments),
		PendingInvestmentCoin: p.PendingInvestmentCoin,
		Alive:                 p.Alive,
	}
}

// GetPrivatePlayerForUpdates retorna a versão privada do jogador para updates
func (p *Player) GetPrivatePlayerForPublicUpdates() PublicPlayerForUpdates {
	privateCardInfos := make([]structs.PublicCardForUpdates, len(p.Cards))
	for i := range p.Cards {
		privateCardInfos[i] = structs.PublicCardForUpdates{
			Name:      &p.Cards[i].Name,
			Index:     p.Cards[i].Index,
			Protected: p.Cards[i].Protected,
			Dead:      p.Cards[i].Dead,
		}
	}

	return PublicPlayerForUpdates{
		Id:                    p.Id,
		Name:                  p.Name,
		MoveTime:              p.MoveTime,
		Cards:                 privateCardInfos,
		Coins:                 p.Coins,
		Investments:           cloneInvestments(p.Investments),
		PendingInvestmentCoin: p.PendingInvestmentCoin,
		Alive:                 p.Alive,
	}
}
