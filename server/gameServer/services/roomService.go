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

func (s *RoomService) ValidatePlayerEntry(r *http.Request, rdb *redis.Client, roomId string, playerId string) (bool, string) {
	//TODO fazer post para o backend para verificar se o playerid existe

	room, err := redisHandlers.LoadRoom(r.Context(), rdb, roomId)
	if err != nil {
		return false, "Sala não encontrada"
	}

	if len(room.Players) >= room.MaxPlayers {
		return false, "Sala está cheia"
	}

	playerInRoom := false
	for _, p := range room.Players {
		if p.Id == playerId {
			playerInRoom = true
			if p.Connected {
				return false, "Player já está na sala"
			}
			// está na sala mas desconectado → pode reconectar
			return true, ""
		}
	}

	if !room.StartTime.IsZero() && !playerInRoom {
		return false, "Jogo já começou"
	}

	roomId, isPlaying, err := redisHandlers.GetRegisteredRoomForPlayer(r.Context(), rdb, playerId)
	if err != nil {
		return false, "Erro ao verificar status do player"
	}
	if isPlaying && roomId != room.ID {
		return false, "Player já está em outra sala"
	}

	return true, ""
}
