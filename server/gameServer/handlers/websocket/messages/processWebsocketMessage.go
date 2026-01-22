package messages

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocols"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocolsValidation"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ProcessPlay(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay, registryRules *rules.Registry) {
	ok, err := roomData.AcquireRoomLock(ctx, rdb, utils.GetInstanceID(), 2*time.Second)
	if err != nil {
		utils.LogError(err)
		return
	}
	if !ok {
		utils.LogDebug("falha ao adquirir lock da sala")
		return
	}
	defer roomData.ReleaseRoomLock(ctx, rdb, utils.GetInstanceID())

	// verifica se o jogador está vivo antes de processar a jogada
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return
	}
	if !sourcePlayer.Alive {
		utils.LogInvldPlyrReq("player %s está morto e não pode jogar", playerPlay.PlayerId)
		return
	}

	// processa o protocolo específico
	switch playerPlay.Type {
	case structs.PlayContest:
		if !protocolsValidation.ValidateContestProtocol(roomData, playerPlay) {
			return
		}
		err := protocols.ContestProtocol(ctx, rdb, registryRules, roomData, playerPlay)
		if err != nil {
			utils.LogError(err)
			return
		}

	case structs.PlayContestPenalty:
		contestPenaltyPayload, ok := playerPlay.Payload.(structs.ContestPenaltyPayload)
		if !ok {
			utils.LogInvldPlyrReq("payload does not match ContestPenaltyPayload structure", sourcePlayer.Id)
			return
		}

		sourcePlayerId, targetPlayerId, valid := protocolsValidation.ValidateContestPenaltyProtocol(roomData, playerPlay, contestPenaltyPayload)
		if !valid {
			return
		}
		err := protocols.ContestPenaltyProtocol(ctx, rdb, registryRules, roomData, contestPenaltyPayload, *sourcePlayerId, *targetPlayerId)
		if err != nil {
			utils.LogError(err)
			return
		}

	case structs.PlayAssassinCard:
		assassinPayload, ok := playerPlay.Payload.(structs.AssassinPayload)
		if !ok {
			utils.LogInvldPlyrReq("payload does not match AssassinPayload structure", sourcePlayer.Id)
			return
		}

		if !protocolsValidation.ValidateAssassinProtocol(roomData, registryRules, playerPlay, assassinPayload) {
			return
		}
		err := protocols.AssassinProtocol(ctx, rdb, registryRules, roomData, playerPlay, assassinPayload)
		if err != nil {
			utils.LogError(err)
			return
		}

	case structs.PlayKamikazeCard:
		kamikazePayload, ok := playerPlay.Payload.(structs.KamikazePayload)
		if !ok {
			utils.LogInvldPlyrReq("payload does not match KamikazePayload structure", sourcePlayer.Id)
			return
		}

		if !protocolsValidation.ValidateKamikazeProtocol(roomData, registryRules, playerPlay, &kamikazePayload) {
			return
		}
		err := protocols.KamikazeProtocol(ctx, rdb, registryRules, roomData, playerPlay, &kamikazePayload)
		if err != nil {
			utils.LogError(err)
			return
		}

	case structs.PlayTrillionaireCard:
		if !protocolsValidation.ValidateTrillionaireProtocol(roomData, registryRules, playerPlay) {
			return
		}
		err := protocols.TrillionaireProtocol(ctx, rdb, registryRules, roomData, playerPlay)
		if err != nil {
			utils.LogError(err)
			return
		}
	}

	// Salva e publica atualizações da sala
	roomData.SaveRoom(ctx, rdb)
	roomData.SendUpdatedRoomData(ctx, rdb)
}
