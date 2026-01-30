package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

func ValidateContestPenaltyProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay, payload structs.ContestPenaltyPayload) (sourcePlayerId *string, targetPlayerId *string, valid bool) {
	// Contest só pode ser usado durante um evento de penalidade de carta por contestação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventContestPenalty {
		utils.LogInvldPlyrReq("Contest Penalty só pode ser usado durante um evento de penalidade de carta por contestação", playerPlay.PlayerId)
		return nil, nil, false
	}

	payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
	if !ok {
		utils.LogError("payload does not match map structure")
		return nil, nil, false
	}

	// Extrair targetPlayer e hasCard do mapa
	contestedPlayerId, ok := payloadMap["contestedPlayer"].(string)
	if !ok {
		utils.LogInvldPlyrReq("contestedPlayer not found or invalid in payload", playerPlay.PlayerId)
		return nil, nil, false
	}
	hasCard, ok := payloadMap["hasCard"].(bool)
	if !ok {
		utils.LogInvldPlyrReq("hasCard not found or invalid in payload", playerPlay.PlayerId)
		return nil, nil, false
	}

	whoWasContestedId := contestedPlayerId

	whoIsContestingId := roomData.GameEvent.PlayerID

	requestPlayerId := playerPlay.PlayerId

	if hasCard {
		if requestPlayerId != whoWasContestedId {
			utils.LogInvldPlyrReq("Apenas o jogador que foi contestado pode usar Contest Penalty quando ele possuia a carta", requestPlayerId)
			return nil, nil, false
		}
		return &whoWasContestedId, &whoIsContestingId, true
	} else {
		if requestPlayerId != whoIsContestingId {
			utils.LogInvldPlyrReq("Apenas o jogador que contestou pode usar Contest Penalty quando ele não possuia a carta", requestPlayerId)
			return nil, nil, false
		}
		return &whoIsContestingId, &whoWasContestedId, true
	}
}
