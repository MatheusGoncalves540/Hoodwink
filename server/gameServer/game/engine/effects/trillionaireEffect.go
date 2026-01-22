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

func TrillionaireEffect(ctx context.Context, rdb *redis.Client, RegistryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	cardRules, err := roomData.GetCardRules(RegistryRules, string(effect.Cause))
	if err != nil {
		utils.LogError(err)
		return
	}

	generalRules, err := roomData.GetGeneralRules(RegistryRules)
	if err != nil {
		utils.LogError(err)
		return
	}

	sourcePlayer, err := roomData.GetPlayer(effect.SourcePlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// calcula o tempo de expiração do efeito
	timeoutDuration, err := roomData.GetTimeoutDuration(RegistryRules, "DisplayMessage")
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	if err != nil {
		return
	}

	// breakLimit indica se o limite de moedas será ultrapassado
	breakLimit := false
	if sourcePlayer.Coins+*cardRules.AmountReceived > *generalRules.MaxCoins {
		breakLimit = true
	}

	// adiciona as moedas ao jogador
	sourcePlayer.AddCoins(*cardRules.AmountReceived)

	// cria o evento de ganho de moedas
	roomData.GameEvent = &structs.GameEvent{
		PlayerID:  effect.SourcePlayer,
		Type:      structs.EventEarnCoins,
		ExpiresAt: expiresAt,
		Payload:   effect.Payload,
	}
	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})

	if breakLimit {
		// se o limite de moedas for ultrapassado, mata a carta viva de menor índice
		aliveIndexes := sourcePlayer.GetAliveCardsIndexes()

		effect := structs.Effect{
			Cause:        structs.EffectGreed,
			SourcePlayer: sourcePlayer.Id,
			Payload: structs.KillCardPayload{
				Cause:           string(structs.EffectGreed),
				TargetPlayer:    &sourcePlayer.Id,
				TargetCardIndex: &aliveIndexes[0],
			},
		}

		KillCard(ctx, rdb, RegistryRules, roomData, effect)
	}
}
