package services

import (
	// "github.com/MatheusGoncalves540/Hoodwink-gameServer/db/models"

	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RoomService struct {
	redisClient *redis.Client
}

func NewRoomService(redisClient *redis.Client) *RoomService {
	return &RoomService{redisClient}
}

func (s *RoomService) CreateNewRoom(r *http.Request, roomData endpointStructures.CreateRoomRequest) (*roomStructs.Room, error) {
	RoomId := uuid.New().String()
	room := &roomStructs.Room{
		ID:                 RoomId,
		Name:               roomData.RoomName,
		Password:           roomData.Password,
		Created:            time.Now(),
		State:              roomStructs.WaitingAction,
		Turn:               0,
		Players:            []roomStructs.Player{},
		MaxPlayers:         roomData.MaxPlayers,
		AliveDeck:          []string{},
		DeadDeck:           []string{},
		CurrentMove:        nil,
		CurrentTurnOwner:   "",
		StartTime:          time.Time{},
		PlayerPending:      "",
		PlayersWhoWantSkip: []string{},
		GameOver:           false,
		PendingEffects:     []roomStructs.Effect{},
	}

	err := redisHandlers.SaveRoom(r.Context(), s.redisClient, room)
	if err != nil {
		return nil, err
	}

	return room, nil
}
