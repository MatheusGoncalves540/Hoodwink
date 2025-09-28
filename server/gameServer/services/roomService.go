package services

import (
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/redisHandlers"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

type RoomService struct {
	redisClient *redis.Client
}

func NewRoomService(redisClient *redis.Client) *RoomService {
	return &RoomService{redisClient}
}

func (s *RoomService) CreateNewRoom(r *http.Request, roomData endpointStructures.CreateRoomRequest) (*roomStructs.Room, error) {
	RoomId := utils.GenerateNewId()
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

	roomTTL := utils.MustEnvInt("ROOM_TTL", 5)
	// Salva a sala com TTL inicial de 5 segundos
	err := redisHandlers.SaveRoomWithTTL(r.Context(), s.redisClient, room, time.Duration(roomTTL)*time.Second)
	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *RoomService) ValidatePlayerEntry(r *http.Request, rdb *redis.Client, backendService *BackendService, roomId string, playerId string) (bool, string) {
	_, err := backendService.GetToBackend("", "/getUserInfoById/"+playerId)
	if err != nil {
		return false, "Erro ao verificar player no backend"
	}

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
