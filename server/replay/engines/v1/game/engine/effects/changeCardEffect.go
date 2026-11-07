package effects

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/pure"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v1/game/structs/rooms"
)

func ChangeCardEffect(ctx context.Context, sink rooms.Sink, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) error {
	// decodifica o payload
	changeCardPayload, ok := effect.Payload.(structs.ChangeCardPayload)
	if !ok {
		pure.LogError("payload inválido para ChangeCardPayload")
		return nil
	}

	// pega o targetPlayer
	targetPlayer, err := roomData.GetPlayer(*changeCardPayload.TargetPlayer)
	if err != nil {
		return err
	}

	// troca a targetCard do targetPlayer
	err = roomData.ChangeCard(targetPlayer, *changeCardPayload.TargetCardIndex)
	if err != nil {
		return err
	}

	if changeCardPayload.UseOnTwoCards != nil && *changeCardPayload.UseOnTwoCards {
		// troca a segunda carta do targetPlayer
		// TODO talvez seja melhor  fazer algo mais genérico para o caso de trocar mais de 2 cartas, mas por enquanto só tem o Croupier que troca 2 cartas
		otherCardIndex := 1 - *changeCardPayload.TargetCardIndex

		err = roomData.ChangeCard(targetPlayer, otherCardIndex)
		if err != nil {
			return err
		}
	}

	return nil
}
