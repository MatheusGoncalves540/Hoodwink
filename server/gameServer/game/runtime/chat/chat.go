package chat

import (
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
)

// ProcessChatMessage anexa a mensagem de chat à sala e devolve aplicado.
// Não persiste nem transmite: o WithRoom salva e o handler transmite.
func ProcessChatMessage(roomData *rooms.Room, playerPlay *structs.PlayerPlay) bool {
	sourcePlayer, err := roomData.GetPlayer(playerPlay.PlayerId)
	if err != nil {
		utils.LogError(err)
		return false
	}

	chatMessagePayload, ok := playerPlay.Payload.(structs.ChatMessagePayload)
	if !ok {
		utils.LogInvldPlyrReq("payload does not match ChatMessagePayload structure", sourcePlayer.Id)
		return false
	}

	roomData.AppendChatMessage(sourcePlayer, chatMessagePayload.Msg)
	return true
}
