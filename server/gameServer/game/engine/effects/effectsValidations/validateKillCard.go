package effectsValidations

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
)

func ValidateKillCardEffect(roomData *roomStructs.Room, effect roomStructs.Effect, payload roomStructs.KillCardPayload) (bool, error) {
	// verifica se o jogador alvo existe
	player, exists := roomData.Players[*payload.TargetPlayer]
	if !exists {
		return false, fmt.Errorf("jogador alvo não encontrado: %s", *payload.TargetPlayer)
	}

	// verifica se o índice da carta é válido
	if *payload.TargetCardIndex <= 0 || *payload.TargetCardIndex > len(player.Cards) {
		return false, fmt.Errorf("índice de carta inválido: %d para jogador %s", *payload.TargetCardIndex, *payload.TargetPlayer)
	}

	return true, nil
}
