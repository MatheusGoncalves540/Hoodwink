package protocols

import (
	"context"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/redis/go-redis/v9"
)

func AssassinProtocol(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, playerPlay *roomStructs.PlayerPlay, assassinPayload roomStructs.AssassinPayload) error {
	cardRules, err := roomData.GetCardRules(registryRules, string(playerPlay.Type))
	if err != nil {
		return fmt.Errorf("%s", "Erro ao obter regras da carta Assassin: "+err.Error())
	}
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		return err
	}

	// calcula o tempo de expiração do evento
	timeoutDuration, err := roomData.GetTimeoutDuration(registryRules, "WaitingAction") // TODO mudar tipo de timeout caso kamikaze esteja ativo na partida
	if err != nil {
		return err
	}
	expiresAt := time.Now().Add(timeoutDuration * time.Second).UTC()

	// marca a remoção de coins
	sourcePlayer.RemoveCoins(*cardRules.Price)

	roomData.GameEvent = &roomStructs.GameEvent{
		PlayerID:  playerPlay.PlayerId,
		Type:      roomStructs.EventCardPlayedAssassin,
		ExpiresAt: expiresAt,
		Payload:   assassinPayload,
	}
	roomData.PendingEffects = append(roomData.PendingEffects,
		roomStructs.Effect{
			Cause:        roomStructs.EffectAssassin,
			SourcePlayer: playerPlay.PlayerId,
			Payload: roomStructs.AssassinPayload{
				TargetPlayer:    assassinPayload.TargetPlayer,
				TargetCardIndex: assassinPayload.TargetCardIndex,
			},
		},
	)

	rdb.ZAdd(ctx, "rooms:timeouts", redis.Z{
		Score:  float64(expiresAt.UnixMilli()),
		Member: roomData.ID,
	})
	return nil
}
