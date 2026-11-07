package rooms

import (
	"github.com/MatheusGoncalves540/Hoodwink-replay/engines/v4/game/structs/players"
)

// GetPublicRoomForUpdates monta a versão pública da sala para updates.
func (r *Room) GetPublicRoomForUpdates() PublicRoomForUpdates {
	roomDataPublic := PublicRoomForUpdates{
		ID:                r.ID,
		MatchID:           r.MatchID,
		Name:              r.Name,
		MaxPlayers:        r.MaxPlayers,
		Turn:              r.Turn,
		Tax:               r.Tax,
		Players:           make(map[string]players.PublicPlayerForUpdates, len(r.Players)),
		Deck:              len(r.Deck),
		CurrentPlayer:     r.CurrentPlayer,
		GameEvent:         r.GameEvent,
		GameOver:          r.GameOver,
		WinnerID:          r.WinnerID,
		StartTime:         r.StartTime,
		DoubledCardValues: r.DoubledCardValues,
		Chat:              r.Chat,
	}

	for playerId, player := range r.Players {
		roomDataPublic.Players[playerId] = player.GetPublicPlayerForUpdates()
	}

	return roomDataPublic
}
