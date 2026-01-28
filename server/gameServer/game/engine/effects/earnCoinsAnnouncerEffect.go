package effects

import (
	"context"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func EarnCoinsAnnouncer(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
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

	// calcula o tempo de expiração do efeito
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "DisplayMessage")
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	if err != nil {
		return err
	}

	// cria o payload de ganho de moedas
	earnCoinsPayload := structs.NewEarnCoinsPayload(string(effect.Cause), EarnCoinsPayload.EarnedCoins, nil)

	// cria o evento de ganho de moedas
	roomData.GameEvent = structs.NewGameEvent(effect.SourcePlayer, structs.EventEarnCoins, expiresAt, earnCoinsPayload)

	// registra o efeito de ganho de moedas
	if err := roomData.AppendPendingEffect(structs.NewEffect(structs.EffectEarnCoins, sourcePlayer.Id, earnCoinsPayload)); err != nil {
		return err
	}

	// registra o timeout do evento
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})

	return nil
}
