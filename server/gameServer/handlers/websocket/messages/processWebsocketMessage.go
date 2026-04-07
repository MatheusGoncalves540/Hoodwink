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
		utils.LogInvldPlyrReq("player "+playerPlay.PlayerId+" está morto e não pode jogar", playerPlay.PlayerId)
		return
	}

	var playHandlers = map[structs.TypePlayerPlays]func() bool{
		structs.PlayContest: func() bool {
			if !protocolsValidation.ValidateContestProtocol(roomData, playerPlay) {
				return false
			}
			if err := protocols.ContestProtocol(ctx, rdb, registryRules, roomData, playerPlay); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayContestPenalty: func() bool {
			contestPenaltyPayload, ok := playerPlay.Payload.(structs.ContestPenaltyPayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match ContestPenaltyPayload structure", sourcePlayer.Id)
				return false
			}

			sourcePlayerId, targetPlayerId, valid := protocolsValidation.ValidateContestPenaltyProtocol(roomData, playerPlay, contestPenaltyPayload)
			if !valid {
				return false
			}
			if err := protocols.ContestPenaltyProtocol(ctx, rdb, registryRules, roomData, contestPenaltyPayload, *sourcePlayerId, *targetPlayerId); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayAssassinCard: func() bool {
			assassinPayload, ok := playerPlay.Payload.(structs.AssassinPayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match AssassinPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateAssassinProtocol(roomData, registryRules, playerPlay, assassinPayload) {
				return false
			}
			if err := protocols.AssassinProtocol(ctx, rdb, registryRules, roomData, playerPlay, assassinPayload); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayKamikazeCard: func() bool {
			kamikazePayload, ok := playerPlay.Payload.(structs.KamikazePayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match KamikazePayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateKamikazeProtocol(roomData, registryRules, playerPlay, &kamikazePayload) {
				return false
			}
			if err := protocols.KamikazeProtocol(ctx, rdb, registryRules, roomData, playerPlay, &kamikazePayload); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayTrillionaireCard: func() bool {
			if !protocolsValidation.ValidateTrillionaireProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.TrillionaireProtocol(ctx, rdb, registryRules, roomData, playerPlay); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayPoliticalCard: func() bool {
			if !protocolsValidation.ValidatePoliticalProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.PoliticalProtocol(ctx, rdb, registryRules, roomData, playerPlay); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayRebelCard: func() bool {
			if !protocolsValidation.ValidateRebelProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.RebelProtocol(ctx, rdb, registryRules, roomData, playerPlay); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayClairvoyantCard: func() bool {
			clairvoyantPayload, ok := playerPlay.Payload.(structs.ClairvoyantPayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match ClairvoyantPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateClairvoyantProtocol(roomData, registryRules, playerPlay, clairvoyantPayload) {
				return false
			}
			if err := protocols.ClairvoyantProtocol(ctx, rdb, registryRules, roomData, playerPlay, clairvoyantPayload); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayGuardianCard: func() bool {
			guardianPayload, ok := playerPlay.Payload.(structs.GuardianPayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match GuardianPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateGuardianProtocol(roomData, registryRules, playerPlay, &guardianPayload) {
				return false
			}
			if err := protocols.GuardianProtocol(ctx, rdb, registryRules, roomData, playerPlay, &guardianPayload); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayTricksterCard: func() bool {
			tricksterPayload, ok := playerPlay.Payload.(structs.TricksterPayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match TricksterPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateTricksterProtocol(roomData, registryRules, playerPlay, &tricksterPayload) {
				return false
			}
			if err := protocols.TricksterProtocol(ctx, rdb, registryRules, roomData, playerPlay, &tricksterPayload); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayGravediggerCard: func() bool {
			gravediggerPayload, ok := playerPlay.Payload.(structs.GravediggerPayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match GravediggerPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateGravediggerProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.GravediggerProtocol(ctx, rdb, registryRules, roomData, playerPlay, &gravediggerPayload); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayCroupierCard: func() bool {
			croupierPayload, ok := playerPlay.Payload.(structs.CroupierPayload)
			if !ok {
				utils.LogInvldPlyrReq("payload does not match CroupierPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateCroupierProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.CroupierProtocol(ctx, rdb, registryRules, roomData, playerPlay, &croupierPayload); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlayInvestorCard: func() bool {
			if !protocolsValidation.ValidateInvestorProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.InvestorProtocol(ctx, rdb, registryRules, roomData, playerPlay); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},

		structs.PlaySelfishCard: func() bool {
			if !protocolsValidation.ValidateSelfishProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.SelfishProtocol(ctx, rdb, registryRules, roomData, playerPlay); err != nil {
				utils.LogError(err)
				return false
			}
			return true
		},
	}

	if handler, ok := playHandlers[playerPlay.Type]; ok {
		if handler() {
			roomData.SaveRoom(ctx, rdb)
			roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{})
		}
	}
}
