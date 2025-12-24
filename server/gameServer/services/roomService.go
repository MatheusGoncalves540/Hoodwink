package services

import (
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/player"
	redisRoom "github.com/MatheusGoncalves540/Hoodwink-gameServer/redis/room"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/endpointStructures"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

type RoomService struct {
	redisClient *redis.Client
	roomTTL     int
}

func NewRoomService(redisClient *redis.Client) *RoomService {
	roomTTL := utils.MustEnvInt("ROOM_TTL", 5)

	return &RoomService{
		redisClient,
		roomTTL,
	}
}

func (s *RoomService) CreateNewRoom(r *http.Request, roomData endpointStructures.CreateRoomRequest, customMatch bool) (*roomStructs.Room, error) {
	RoomId := utils.GenerateNewId()
	room := &roomStructs.Room{
		ID:            RoomId,
		Name:          roomData.RoomName,
		Password:      roomData.Password,
		MaxPlayers:    roomData.MaxPlayers,
		CustomMatch:   customMatch,
		Turn:          1,
		Tax:           0,
		Players:       make(map[string]roomStructs.Player),
		DeadDeck:      []string{},
		CurrentPlayer: "",
		State:         roomStructs.StateWaitingFirstAction,
		PendingEvent:  &roomStructs.PendingEvent{},
		PendingPlayer: "",
		GameOver:      false,
		StartTime:     time.Time{},
		Created:       time.Now(),
	}

	// Salva a sala com TTL inicial de 5 segundos
	err := redisRoom.SaveRoomWithTTL(r.Context(), s.redisClient, room, time.Duration(s.roomTTL)*time.Second)
	if err != nil {
		return nil, err
	}

	// TODO transformar isso em "waiting game start"
	engine.WaitingFirstAction(r.Context(), s.redisClient, room)

	return room, nil
}

func (s *RoomService) ValidatePlayerEntry(r *http.Request, rdb *redis.Client, backendService *BackendService, roomId string, playerId string) (bool, string) {
	_, err := backendService.GetToBackend("/getUserInfoById/" + playerId)
	if err != nil {
		return false, "Erro ao verificar player no backend"
	}

	room, err := redisRoom.LoadRoom(r.Context(), rdb, roomId)
	if err != nil {
		return false, "Sala não encontrada"
	}

	_, playerInRoom := room.Players[playerId]

	if !playerInRoom && len(room.Players) >= room.MaxPlayers {
		return false, "Sala está cheia"
	}

	if !room.StartTime.IsZero() && !playerInRoom {
		return false, "Jogo já começou"
	}

	roomId, isPlaying, err := player.GetRegisteredRoomForPlayer(r.Context(), rdb, playerId)
	if err != nil {
		return false, "Erro ao verificar status do player"
	}
	if isPlaying && roomId != room.ID {
		return false, "Player já está em outra sala"
	}

	return true, ""
}

func (s *RoomService) SyncActivePlayers(r *http.Request, rdb *redis.Client, roomId string) error {
	room, err := redisRoom.LoadRoom(r.Context(), rdb, roomId)
	if err != nil {
		return err
	}

	// Itera sobre os players e remove aqueles sem registro ativo
	for playerId := range room.Players {
		registeredRoom, registered, err := player.GetRegisteredRoomForPlayer(r.Context(), rdb, playerId)
		if err != nil {
			return err
		}
		if !registered || registeredRoom != roomId {
			delete(room.Players, playerId)
		}
	}

	// Salva a sala atualizada
	return redisRoom.SaveRoom(r.Context(), rdb, room)
}
