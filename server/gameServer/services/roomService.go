package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/engine"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/roomStructs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
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

func (s *RoomService) CreateNewRoom(r *http.Request, roomData endpointStructures.CreateRoomRequest, customMatch bool) (*rooms.Room, error) {
	RoomId := utils.GenerateNewId()
	// TODO: desmockar tudo isso
	room := &rooms.Room{
		ID:            RoomId,
		Rules:         roomStructs.ClassicRules,
		TimeoutType:   string(rules.TimeoutsTypeDefault),
		Name:          roomData.RoomName,
		Password:      roomData.Password,
		MaxPlayers:    roomData.MaxPlayers,
		CustomMatch:   customMatch,
		Turn:          1,
		Tax:           0,
		Players:       make(map[string]players.Player),
		DeadDeck:      []string{},
		CurrentPlayer: "",
		GameEvent:     &roomStructs.GameEvent{},
		PendingPlayer: "",
		GameOver:      false,
		StartTime:     time.Time{},
		Created:       time.Now(),
	}

	// Salva a sala com TTL inicial de 5 segundos
	err := room.SaveRoomWithTTL(r.Context(), s.redisClient, time.Duration(s.roomTTL)*time.Second)
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

	roomData, err := redisFuncs.LoadRoom(r.Context(), rdb, roomId)
	if err != nil {
		return false, "Sala não encontrada"
	}

	player, err := roomData.GetPlayer(playerId)
	if err != nil {
		return false, "Erro ao verificar player"
	}

	if player == nil && len(roomData.Players) >= roomData.MaxPlayers {
		return false, "Sala está cheia"
	}

	if !roomData.StartTime.IsZero() && player == nil {
		return false, "Jogo já começou"
	}

	roomId, isPlaying, err := player.GetRegisteredRoomForPlayer(r.Context(), rdb)
	if err != nil {
		return false, "Erro ao verificar status do player"
	}
	if isPlaying && roomId != roomData.ID {
		return false, "Player já está em outra sala"
	}

	return true, ""
}

func (s *RoomService) SyncActivePlayers(r *http.Request, rdb *redis.Client, roomId string) error {
	roomData, err := redisFuncs.LoadRoom(r.Context(), rdb, roomId)
	if err != nil {
		return err
	}

	// Itera sobre os players e remove aqueles sem registro ativo
	for playerId := range roomData.Players {
		player, err := roomData.GetPlayer(playerId)
		if err != nil {
			return err
		}
		if player == nil {
			return fmt.Errorf("jogador %s não encontrado na sala %s, sendo que devia estar", playerId, roomId)
		}

		registeredRoom, registered, err := player.GetRegisteredRoomForPlayer(r.Context(), rdb)
		if err != nil {
			return err
		}
		if !registered || registeredRoom != roomId {
			delete(roomData.Players, playerId)
		}
	}

	// Salva a sala atualizada
	return roomData.SaveRoom(r.Context(), rdb)
}
