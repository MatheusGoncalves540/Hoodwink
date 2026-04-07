package players

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
)

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
		Id:    p.Id,
		Name:  p.Name,
		Cards: privateCardInfos,
		Coins: p.Coins,
		Alive: p.Alive,
	}
}
