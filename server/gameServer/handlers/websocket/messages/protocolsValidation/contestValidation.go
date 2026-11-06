package protocolsValidation

import (
	"strings"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateContestProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	// Contest só pode ser usado durante um evento de carta
	if roomData.GameEvent == nil || !strings.Contains(string(roomData.GameEvent.Type), "CARD_PLAYED") {
		utils.LogInvldPlyrReq("Contest só pode ser usado durante um evento de carta", playerPlay.PlayerId)
		return false
	}

	// Extrair o player que usou a carta sendo contestada
	playerThatUseTheCard := roomData.GameEvent.PlayerID

	// verifica se o player que usou a carta foi extraído corretamente
	if playerThatUseTheCard == "" {
		utils.LogInvldPlyrReq("PlayerID not found or invalid in GameEvent", playerPlay.PlayerId)
		return false
	}

	// verifica se o jogador que está tentando contestar não é ele mesmo
	if playerThatUseTheCard == playerPlay.PlayerId {
		utils.LogInvldPlyrReq("Player tentou contestar um evento que dele próprio", playerPlay.PlayerId)
		return false
	}

	return true
}
