package protocolsValidation

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/rooms"
)

// ValidateContestPenaltyProtocol valida a escolha da carta que morre na contestação
// e devolve o payload normalizado com os dados calculados pelo servidor.
func ValidateContestPenaltyProtocol(roomData *rooms.Room, playerPlay *structs.PlayerPlay, payload structs.ContestPenaltyPayload) (*structs.ContestPenaltyPayload, *string, *string, bool) {
	// Contest Penalty só pode ser usado durante um evento de penalidade de contestação
	if roomData.GameEvent == nil || roomData.GameEvent.Type != structs.EventContestPenalty {
		pure.LogInvldPlyrReq("Contest Penalty só pode ser usado durante um evento de penalidade de carta por contestação", playerPlay.PlayerId)
		return nil, nil, nil, false
	}

	payloadMap, ok := roomData.GameEvent.Payload.(map[string]interface{})
	if !ok {
		pure.LogError("payload does not match map structure")
		return nil, nil, nil, false
	}

	contestedPlayerId, ok := payloadMap["contestedPlayer"].(string)
	if !ok {
		pure.LogInvldPlyrReq("contestedPlayer not found or invalid in payload", playerPlay.PlayerId)
		return nil, nil, nil, false
	}
	contestingPlayerId, ok := payloadMap["contestingPlayer"].(string)
	if !ok {
		pure.LogInvldPlyrReq("contestingPlayer not found or invalid in payload", playerPlay.PlayerId)
		return nil, nil, nil, false
	}
	contestedCard, ok := payloadMap["contestedCard"].(string)
	if !ok {
		pure.LogInvldPlyrReq("contestedCard not found or invalid in payload", playerPlay.PlayerId)
		return nil, nil, nil, false
	}
	hasCard, ok := payloadMap["hasCard"].(bool)
	if !ok {
		pure.LogInvldPlyrReq("hasCard not found or invalid in payload", playerPlay.PlayerId)
		return nil, nil, nil, false
	}

	chooserID := contestingPlayerId
	targetPlayerID := contestedPlayerId
	if hasCard {
		chooserID = contestedPlayerId
		targetPlayerID = contestingPlayerId
	}

	if playerPlay.PlayerId != chooserID {
		pure.LogInvldPlyrReq("Apenas o jogador responsável pela escolha pode usar Contest Penalty", playerPlay.PlayerId)
		return nil, nil, nil, false
	}

	if payload.TargetCardIndex == nil {
		pure.LogInvldPlyrReq("targetCardIndex é obrigatório", playerPlay.PlayerId)
		return nil, nil, nil, false
	}

	targetPlayer, err := roomData.GetPlayer(targetPlayerID)
	if err != nil {
		pure.LogError(err)
		return nil, nil, nil, false
	}
	card, err := targetPlayer.GetCardByIndex(*payload.TargetCardIndex)
	if err != nil || card.Dead {
		pure.LogInvldPlyrReq("targetCardIndex não aponta para uma carta viva do alvo", playerPlay.PlayerId)
		return nil, nil, nil, false
	}

	normalized := structs.NewContestPenaltyPayload(
		contestedPlayerId,
		contestingPlayerId,
		contestedCard,
		hasCard,
		payload.TargetCardIndex,
	)

	// O índice da carta revelada é sempre calculado pelo servidor.
	if revealedCardIndex, ok := payloadMap["revealedCardIndex"].(float64); ok {
		index := int(revealedCardIndex)
		normalized.RevealedCardIndex = &index
	}

	return &normalized, &chooserID, &targetPlayerID, true
}
