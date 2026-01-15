package effectsValidations

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/players"
)

// aqui não preisa verificar as mesmas coisas do efeito de killCard, isso já foi verificado na resolução do efeito
func ValidateKamikaze(player *players.Player) (bool, error) {
	// verifica se o jogador está vivo
	// if !player.Alive {
	// 	return false, fmt.Errorf("jogador %s não está vivo para usar Kamikaze", player.Id)
	// }

	return true, nil
}
