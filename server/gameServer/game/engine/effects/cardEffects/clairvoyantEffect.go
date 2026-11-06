package cardEffects

import (
	"context"
	"slices"

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

	// verifica se a carta já está protegida, se estiver, não pode ser revelada, apenas remove a proteção
	if slices.Contains(targetPlayer.GetProtectedCardsIndexes(), clairvoyantPayload.TargetCardIndex) {
		err = targetPlayer.UnprotectCard(clairvoyantPayload.TargetCardIndex)
		if err != nil {
			utils.LogError(err)
		}
		return
	}

	// pega a carta revelada alvo
	revealedCard, err := targetPlayer.GetCardByIndex(clairvoyantPayload.TargetCardIndex)
	if err != nil {
		utils.LogError(err)
		return
	}
	if revealedCard.Dead {
		utils.LogInvldPlyrReq("Clairvoyant tentou revelar uma carta morta", effect.SourcePlayer)
		return
	}
	revealedCardName := string(revealedCard.Name)

	// A carta é revelada momentaneamente e depois escondida novamente, sem trocar de posição.
	// A revelação é feita pelo evento de apresentação, não alterando o estado permanente da carta.
	clairvoyantPayload.RevealedCard = &revealedCardName

	if clairvoyantPayload.ShowToAllPlayers {
		if err := roomData.AppendPendingPresentationEvent(structs.NewPresentationEvent(effect.SourcePlayer, structs.EventRevealedCard, clairvoyantPayload)); err != nil {
			utils.LogError(err)
			return
		}
		return
	}

	presentationEvent := structs.NewPresentationEvent(effect.SourcePlayer, structs.EventRevealedCard, structs.ClairvoyantPayload{
		TargetPlayer:     clairvoyantPayload.TargetPlayer,
		TargetCardIndex:  clairvoyantPayload.TargetCardIndex,
		ShowToAllPlayers: false,
	})
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
}
