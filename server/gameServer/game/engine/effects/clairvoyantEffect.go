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

func ClairvoyantEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) (*rooms.Room, []string) {
	clairvoyantPayload, ok := effect.Payload.(structs.ClairvoyantPayload)
	if !ok {
		utils.LogError("payload inválido para EffectClairvoyant")
		return nil, nil
	}

	targetPlayer, err := roomData.GetPlayer(clairvoyantPayload.TargetPlayer)
	if err != nil {
		utils.LogError(err)
		return nil, nil
	}

	// calcula o tempo de expiração do efeito
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "DisplayImportantMessage")
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()
	if err != nil {
		utils.LogError(err)
		return nil, nil
	}

	// pega a carta revelada alvo
	revealedCard, err := targetPlayer.GetCardByIndex(clairvoyantPayload.TargetCardIndex)
	if err != nil {
		utils.LogError(err)
		return nil, nil
	}
	revealedCardName := string(revealedCard.Name)

	if clairvoyantPayload.ShowToAllPlayers {
		// adiciona o nome da carta revelada ao payload
		clairvoyantPayload.RevealedCard = &revealedCardName

		roomData.GameEvent = structs.NewGameEvent(effect.SourcePlayer, structs.EventRevealedCard, expiresAt, clairvoyantPayload)

		rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
			Score:  float64(expiresAt.UnixMilli()),
			Member: roomData.ID,
		})

		return nil, nil
	} else {
		roomData.GameEvent = structs.NewGameEvent(effect.SourcePlayer, structs.EventRevealedCard, expiresAt, clairvoyantPayload)

		confidencialRoomData := roomData.Clone()
		playersThatCanSee := []string{}

		confidencialRoomData.GameEvent.Payload = structs.ClairvoyantPayload{
			TargetPlayer:     clairvoyantPayload.TargetPlayer,
			TargetCardIndex:  clairvoyantPayload.TargetCardIndex,
			ShowToAllPlayers: false,
			RevealedCard:     &revealedCardName,
		}

		playersThatCanSee = append(playersThatCanSee, effect.SourcePlayer)

		rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
			Score:  float64(expiresAt.UnixMilli()),
			Member: roomData.ID,
		})

		return confidencialRoomData, playersThatCanSee
	}
}
