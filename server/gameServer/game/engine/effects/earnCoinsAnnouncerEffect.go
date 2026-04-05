package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func EarnCoinsAnnouncer(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	EarnCoinsPayload, ok := effect.Payload.(structs.EarnCoinsPayload)
	if !ok {
		utils.LogError("payload inválido para KillCardPayload")
		return nil
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		return err
	}

	// cria o payload de ganho de moedas
	earnCoinsPayload := structs.NewEarnCoinsPayload(string(effect.Cause), EarnCoinsPayload.EarnedCoins, nil)

	if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventEarnCoins, earnCoinsPayload)); err != nil {
		return err
	}

	// registra o efeito de ganho de moedas
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectEarnCoins, sourcePlayer.Id, earnCoinsPayload)); err != nil {
		return err
	}

	return nil
}
