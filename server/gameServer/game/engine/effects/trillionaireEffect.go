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

	earnedCoins, err := roomData.GetCardValue(RegistryRules, structs.TypePlayerPlays(effect.Cause))
	if err != nil {
		utils.LogError(err)
		return
	}

	// adiciona as moedas ao jogador e retorna breakLimit, que indica se o limite de moedas será ultrapassado
	breakLimit := sourcePlayer.AddCoins(earnedCoins, *generalRules.MaxCoins)

	if breakLimit {
		// se o limite de moedas for ultrapassado, mata a carta viva de menor índice
		aliveIndexes := sourcePlayer.GetAliveCardsIndexes()

		greedPayload := structs.NewKillCardPayload(string(structs.EffectGreed), &sourcePlayer.Id, &aliveIndexes[0])

		effect := structs.Effect{
			Cause:        structs.EffectGreed,
			SourcePlayer: sourcePlayer.Id,
			Payload:      greedPayload,
		}

		KillCard(ctx, rdb, RegistryRules, roomData, effect)
	} else {
		// cria o evento de ganho de moedas
		roomData.GameEvent = structs.NewGameEvent(effect.SourcePlayer, structs.EventEarnCoins, expiresAt, effect.Payload)

		// registra o timeout do evento
		rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
			Score:  float64(expiresAt.UnixMilli()),
			Member: roomData.ID,
		})
	}
}
