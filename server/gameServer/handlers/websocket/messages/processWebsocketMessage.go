package messages

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/config"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocols"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/handlers/websocket/messages/protocolsValidation"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ProcessPlay(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay, registryRules *rules.Registry) {
	ok, err := roomData.AcquireRoomLock(ctx, rdb, config.InstanceID, 2*time.Second)
	if err != nil {
		utils.LogError(err)
		return
	}
	if !ok {
		utils.LogDebug("falha ao adquirir lock da sala")
		return
	}
	defer roomData.ReleaseRoomLock(ctx, rdb, config.InstanceID)

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
	case structs.PlayAssassinCard:
		assassinPayload, ok := playerPlay.Payload.(structs.AssassinPayload)
		if !ok {
			utils.LogInvldPlyrReq("payload does not match AssassinPayload structure", sourcePlayer.Id)
			return
		}

		if protocolsValidation.ValidateAssassinProtocol(roomData, registryRules, playerPlay, assassinPayload) {
			protocols.AssassinProtocol(ctx, rdb, registryRules, roomData, playerPlay, assassinPayload)
		}

	case structs.PlayKamikazeCard:
		kamikazePayload, ok := playerPlay.Payload.(structs.KamikazePayload)
		if !ok {
			utils.LogInvldPlyrReq("payload does not match KamikazePayload structure", sourcePlayer.Id)
			return
		}

		if protocolsValidation.ValidateKamikazeProtocol(roomData, registryRules, playerPlay, &kamikazePayload) {
			protocols.KamikazeProtocol(ctx, rdb, registryRules, roomData, playerPlay, &kamikazePayload)
		}
	}
}
