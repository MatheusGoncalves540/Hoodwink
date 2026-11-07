package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v3/game/structs/rooms"
)

// ValidatePassProtocol aceita "deixar passar" só em janelas de reação e só de quem
// não é o autor da jogada que abriu a janela.
func ValidatePassProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	if !roomData.IsReactionWindow() {
		pure.LogInvldPlyrReq("PASS só pode ser usado durante uma janela de reação", playerPlay.PlayerId)
		return false
	}

	if roomData.GameEvent.PlayerID == playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou deixar passar a própria jogada", playerPlay.PlayerId)
		return false
	}

	return true
}
