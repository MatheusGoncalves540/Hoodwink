package protocolsValidation

import (
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v5/game/structs/rooms"
)

func ValidateContestProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	// Contest só pode ser usado durante um evento de carta
	if roomData.GameEvent == nil || !strings.Contains(string(roomData.GameEvent.Type), "CARD_PLAYED") {
		pure.LogInvldPlyrReq("Contest só pode ser usado durante um evento de carta", playerPlay.PlayerId)
		return false
	}

	// Extrair o player que usou a carta sendo contestada
	playerThatUseTheCard := roomData.GameEvent.PlayerID

	// verifica se o player que usou a carta foi extraído corretamente
	if playerThatUseTheCard == "" {
		pure.LogInvldPlyrReq("PlayerID not found or invalid in GameEvent", playerPlay.PlayerId)
		return false
	}

	// verifica se o jogador que está tentando contestar não é ele mesmo
	if playerThatUseTheCard == playerPlay.PlayerId {
		pure.LogInvldPlyrReq("Player tentou contestar um evento que dele próprio", playerPlay.PlayerId)
		return false
	}

	return true
}
