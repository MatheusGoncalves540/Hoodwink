package engine

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/handlers/websocket/messages/protocols"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v2/handlers/websocket/messages/protocolsValidation"
)

// ProcessPlay aplica a jogada e, quando aceita, transmite via Sink. A persistência
// fica com o WithRoom (que salva quando o fn devolve aplicado).
func ProcessPlay(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, playerPlay *structs.PlayerPlay, registryRules *rules.Registry) bool {
	sink = rooms.SinkOrNoop(sink)

	if ApplyPlay(ctx, sink, roomData, playerPlay, registryRules) {
		sink.Broadcast(ctx, roomData, nil, []string{})
		return true
	}
	return false
}

// ApplyPlay valida e aplica a jogada na sala em memória; true quando foi aceita.
// Não persiste nem transmite: ProcessPlay faz isso, e os testes de cenário rodam sem Redis.
func ApplyPlay(ctx context.Context, sink rooms.Sink, roomData *rooms.Room, playerPlay *structs.PlayerPlay, registryRules *rules.Registry) bool {
	sink = rooms.SinkOrNoop(sink)

	if roomData.GameOver {
		pure.LogInvldPlyrReq("partida encerrada", playerPlay.PlayerId)
		return false
	}

	// verifica se o jogador está vivo antes de processar a jogada
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		pure.LogError(err)
		return false
	}
	if !sourcePlayer.Alive {
		pure.LogInvldPlyrReq("player "+playerPlay.PlayerId+" está morto e não pode jogar", playerPlay.PlayerId)
		return false
	}

	var playHandlers = map[structs.TypePlayerPlays]func() bool{
		structs.PlayContest: func() bool {
			if !protocolsValidation.ValidateContestProtocol(roomData, playerPlay) {
				return false
			}
			if err := protocols.ContestProtocol(ctx, sink, registryRules, roomData, playerPlay); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayContestPenalty: func() bool {
			contestPenaltyPayload, ok := playerPlay.Payload.(structs.ContestPenaltyPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match ContestPenaltyPayload structure", sourcePlayer.Id)
				return false
			}

			normalizedPayload, chooserID, _, valid := protocolsValidation.ValidateContestPenaltyProtocol(roomData, playerPlay, contestPenaltyPayload)
			if !valid {
				return false
			}
			if err := protocols.ContestPenaltyProtocol(ctx, sink, registryRules, roomData, *normalizedPayload, *chooserID); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayTakeCoins: func() bool {
			if !protocolsValidation.ValidateTakeCoinsProtocol(roomData, playerPlay) {
				return false
			}
			if err := protocols.TakeCoinsProtocol(ctx, sink, registryRules, roomData, playerPlay); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayBasicKill: func() bool {
			killPayload, ok := playerPlay.Payload.(structs.AssassinPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match AssassinPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateBasicKillProtocol(roomData, registryRules, playerPlay, killPayload) {
				return false
			}
			if err := protocols.BasicKillProtocol(ctx, sink, registryRules, roomData, playerPlay, killPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlaySwapCards: func() bool {
			swapPayload, ok := playerPlay.Payload.(structs.SwapCardsPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match SwapCardsPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateSwapCardsProtocol(roomData, registryRules, playerPlay, swapPayload) {
				return false
			}
			if err := protocols.SwapCardsProtocol(ctx, sink, registryRules, roomData, playerPlay, swapPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayAssassinCard: func() bool {
			assassinPayload, ok := playerPlay.Payload.(structs.AssassinPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match AssassinPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateAssassinProtocol(roomData, registryRules, playerPlay, assassinPayload) {
				return false
			}
			if err := protocols.AssassinProtocol(ctx, sink, registryRules, roomData, playerPlay, assassinPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayKamikazeCard: func() bool {
			kamikazePayload, ok := playerPlay.Payload.(structs.KamikazePayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match KamikazePayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateKamikazeProtocol(roomData, registryRules, playerPlay, &kamikazePayload) {
				return false
			}
			if err := protocols.KamikazeProtocol(ctx, sink, registryRules, roomData, playerPlay, &kamikazePayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayTrillionaireCard: func() bool {
			if !protocolsValidation.ValidateTrillionaireProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.TrillionaireProtocol(ctx, sink, registryRules, roomData, playerPlay); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayPoliticalCard: func() bool {
			if !protocolsValidation.ValidatePoliticalProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.PoliticalProtocol(ctx, sink, registryRules, roomData, playerPlay); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayRebelCard: func() bool {
			if !protocolsValidation.ValidateRebelProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.RebelProtocol(ctx, sink, registryRules, roomData, playerPlay); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayClairvoyantCard: func() bool {
			clairvoyantPayload, ok := playerPlay.Payload.(structs.ClairvoyantPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match ClairvoyantPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateClairvoyantProtocol(roomData, registryRules, playerPlay, clairvoyantPayload) {
				return false
			}
			if err := protocols.ClairvoyantProtocol(ctx, sink, registryRules, roomData, playerPlay, clairvoyantPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayGuardianCard: func() bool {
			guardianPayload, ok := playerPlay.Payload.(structs.GuardianPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match GuardianPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateGuardianProtocol(roomData, registryRules, playerPlay, &guardianPayload) {
				return false
			}
			if err := protocols.GuardianProtocol(ctx, sink, registryRules, roomData, playerPlay, &guardianPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayTricksterCard: func() bool {
			tricksterPayload, ok := playerPlay.Payload.(structs.TricksterPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match TricksterPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateTricksterProtocol(roomData, registryRules, playerPlay, &tricksterPayload) {
				return false
			}
			if err := protocols.TricksterProtocol(ctx, sink, registryRules, roomData, playerPlay, &tricksterPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayGravediggerCard: func() bool {
			gravediggerPayload, ok := playerPlay.Payload.(structs.GravediggerPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match GravediggerPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateGravediggerProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.GravediggerProtocol(ctx, sink, registryRules, roomData, playerPlay, &gravediggerPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayCroupierCard: func() bool {
			croupierPayload, ok := playerPlay.Payload.(structs.CroupierPayload)
			if !ok {
				pure.LogInvldPlyrReq("payload does not match CroupierPayload structure", sourcePlayer.Id)
				return false
			}

			if !protocolsValidation.ValidateCroupierProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.CroupierProtocol(ctx, sink, registryRules, roomData, playerPlay, &croupierPayload); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlayInvestorCard: func() bool {
			if !protocolsValidation.ValidateInvestorProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.InvestorProtocol(ctx, sink, registryRules, roomData, playerPlay); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},

		structs.PlaySelfishCard: func() bool {
			if !protocolsValidation.ValidateSelfishProtocol(roomData, registryRules, playerPlay) {
				return false
			}
			if err := protocols.SelfishProtocol(ctx, sink, registryRules, roomData, playerPlay); err != nil {
				pure.LogError(err)
				return false
			}
			return true
		},
	}

	handler, ok := playHandlers[playerPlay.Type]
	return ok && handler()
}
