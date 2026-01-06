package effectsValidations

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
)

func ValidateKillCardEffect(roomData *roomStructs.Room, effect roomStructs.Effect, payload roomStructs.KillCardPayload, player *roomStructs.Player) (bool, error) {
	// verifica se o índice da carta é válido
	if *payload.TargetCardIndex < 0 || *payload.TargetCardIndex >= len(player.Cards) {
		return false, fmt.Errorf("índice de carta inválido: %d para jogador %s", *payload.TargetCardIndex, *payload.TargetPlayer)
	}

	// verifica se a carta já está morta
	card, err := player.GetCardByIndex(*payload.TargetCardIndex)
	if err != nil {
		return false, err
	}
	if card.Dead {
		return false, fmt.Errorf("carta já está morta: índice %d do jogador %s", *payload.TargetCardIndex, *payload.TargetPlayer)
	}

	return true, nil
}
