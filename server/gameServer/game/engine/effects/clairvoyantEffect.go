package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ClairvoyantEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) {
	clairvoyantPayload, ok := effect.Payload.(structs.ClairvoyantPayload)
	if !ok {
		utils.LogError("payload inválido para EffectClairvoyant")
		return
	}

	targetPlayer, err := roomData.GetPlayer(clairvoyantPayload.TargetPlayer)
	if err != nil {
		utils.LogError(err)
		return
	}

	// pega a carta revelada alvo
	revealedCard, err := targetPlayer.GetCardByIndex(clairvoyantPayload.TargetCardIndex)
	if err != nil {
		utils.LogError(err)
		return
	}
	revealedCardName := string(revealedCard.Name)

	if clairvoyantPayload.ShowToAllPlayers {
		// adiciona o nome da carta revelada ao payload
		clairvoyantPayload.RevealedCard = &revealedCardName

		if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventRevealedCard, clairvoyantPayload)); err != nil {
			utils.LogError(err)
			return
		}

		return
	} else {
		presentationEvent := structs.NewPresentationEvent(effect.SourcePlayer, structs.EventRevealedCard, clairvoyantPayload)
		presentationEvent.ConfidencialPayload = structs.ClairvoyantPayload{
			TargetPlayer:     clairvoyantPayload.TargetPlayer,
			TargetCardIndex:  clairvoyantPayload.TargetCardIndex,
			ShowToAllPlayers: false,
			RevealedCard:     &revealedCardName,
		}
		presentationEvent.ConfidencialPlayerIds = []string{effect.SourcePlayer}

		if err := roomData.AppendPendingPresentationEvent(presentationEvent); err != nil {
			utils.LogError(err)
			return
		}

		return
	}
}
