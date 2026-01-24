package effectsValidations

import (
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

func ValidateKillCardEffect(roomData *rooms.Room, effect structs.Effect, payload structs.KillCardPayload, targetPlayer *players.Player) (bool, error) {
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
		// escolhe a carta com menor índice que não esteja morta
		aliveIndexes := targetPlayer.GetAliveCardsIndexes()
		if len(aliveIndexes) == 0 {
			return false, fmt.Errorf("todas as cartas do jogador %s estão mortas", *payload.TargetPlayer)
		}
		*payload.TargetCardIndex = aliveIndexes[0]
	}

	return true, nil
}
