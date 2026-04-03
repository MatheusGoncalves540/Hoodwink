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
