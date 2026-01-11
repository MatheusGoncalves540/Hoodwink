package effectsValidations

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
)

// aqui não preisa verificar as mesmas coisas do efeito de killCard, isso já foi verificado na resolução do efeito
func ValidateAssassin(roomData *rooms.Room, cardRules *rules.CardRules, effect roomStructs.Effect, player *players.Player) (bool, error) {
	// verifica se o jogador tem moedas suficientes
	if player.Coins < *cardRules.Price {
		return false, fmt.Errorf("jogador %s não tem moedas suficientes para usar Assassin", player.Id)
	}

	return true, nil
}
