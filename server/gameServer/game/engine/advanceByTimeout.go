package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

// AdvanceByTimeout avança o estado do jogo baseado no estado atual da sala.
func AdvanceByTimeout(ctx context.Context, rdb *redis.Client, room *roomStructs.Room) {
	switch room.State {

	case roomStructs.StateWaitingFirstAction:
		if err := NextTurn(room, rdb, ctx); err != nil {
			utils.LogError(err)
		}
		// TODO: Remover depois de testar
		data, _ := json.MarshalIndent(room, "", "  ")
		fmt.Println(string(data))

	case roomStructs.StateWaitAction:
		// AdvanceTurn(room)

	case roomStructs.StateWaitContest:
		// ApplyAction(room)
		// StartKamikaze(room)
	}
}
