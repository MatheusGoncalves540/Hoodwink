package chat

import (
	"context"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

func ProcessChatMessage(ctx context.Context, rdb *redis.Client, roomData *rooms.Room, playerPlay *structs.PlayerPlay) {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return
	}

	chatMessagePayload, ok := playerPlay.Payload.(structs.ChatMessagePayload)
	if !ok {
		utils.LogInvldPlyrReq("payload does not match ChatMessagePayload structure", sourcePlayer.Id)
		return
	}

	roomData.AppendChatMessage(sourcePlayer, chatMessagePayload.Msg)

	if err := roomData.SaveRoom(ctx, rdb); err != nil {
		utils.LogError(err)
		return
	}
}
