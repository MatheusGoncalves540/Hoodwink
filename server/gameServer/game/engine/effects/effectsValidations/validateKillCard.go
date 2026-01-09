package effectsValidations

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
)

func ValidateKillCardEffect(roomData *roomStructs.Room, effect roomStructs.Effect, payload roomStructs.KillCardPayload, targetPlayer *roomStructs.Player) (bool, error) {
	// verifica se o jogador que está tentando matar a carta é o mesmo que a carta pertence
	if targetPlayer.Id == effect.SourcePlayer {
		return false, fmt.Errorf("jogador não pode matar a si mesmo")
	}

	// verifica se o índice da carta é válido
	if *payload.TargetCardIndex < 0 || *payload.TargetCardIndex >= len(targetPlayer.Cards) {
		return false, fmt.Errorf("índice de carta inválido: %d para jogador %s", *payload.TargetCardIndex, *payload.TargetPlayer)
	}

	// verifica se a carta já está morta
	card, err := targetPlayer.GetCardByIndex(*payload.TargetCardIndex)
	if err != nil {
		return false, err
	}
	if card.Dead {
		return false, fmt.Errorf("carta já está morta: índice %d do jogador %s", *payload.TargetCardIndex, *payload.TargetPlayer)
	}

	return true, nil
}
