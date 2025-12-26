package effectsValidations

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateKillCardEffect(roomData *roomStructs.Room, effect roomStructs.Effect, payload roomStructs.KillCardPayload) (bool, error) {
	// verifica se o jogador alvo existe
	player, exists := roomData.Players[*payload.TargetPlayer]
	if !exists {
		utils.LogError(fmt.Errorf("jogador alvo não encontrado: %s", *payload.TargetPlayer))
		return false, fmt.Errorf("jogador alvo não encontrado: %s", *payload.TargetPlayer)
	}

	// verifica se o índice da carta é válido
	if *payload.TargetCard < 0 || *payload.TargetCard >= len(player.Cards) {
		utils.LogError(fmt.Errorf("índice de carta inválido: %d para jogador %s", *payload.TargetCard, *payload.TargetPlayer))
		return false, fmt.Errorf("índice de carta inválido: %d para jogador %s", *payload.TargetCard, *payload.TargetPlayer)
	}

	return true, nil
}
