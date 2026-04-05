package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
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

func (s *RoomService) CreateNewRoom(r *http.Request, registryRules *rules.Registry, createRoomRequest endpointStructures.CreateRoomRequest, customMatch bool) (*rooms.Room, error) {
	RoomId := utils.GenerateNewId()
	// TODO: desmockar tudo isso
	roomData := &rooms.Room{
		ID:                        RoomId,
		Rules:                     structs.ClassicRules,
		TimeoutType:               string(rules.TimeoutsTypeDefault),
		Name:                      createRoomRequest.RoomName,
		Password:                  createRoomRequest.Password,
		MaxPlayers:                createRoomRequest.MaxPlayers,
		CustomMatch:               customMatch,
		Turn:                      1,
		Tax:                       0,
		DoubledCardValues:         make(map[structs.TypePlayerPlays]structs.DoubledCardValues),
		Players:                   make(map[string]*players.Player),
		Deck:                      make([]structs.CardName, 0),
		CurrentPlayer:             "",
		GameEvent:                 structs.NewGameEvent("", "", time.Time{}, nil),
		PendingEffects:            []structs.EffectDTO{},
		PendingPresentationEvents: []structs.PresentationEventDTO{},
		GameOver:                  false,
		StartTime:                 time.Time{},
		Created:                   time.Now().UTC(),
	}

	// obtém as regras gerais do jogo
	generalRules, err := roomData.GetGeneralRules(registryRules)
	if err != nil {
		utils.LogError(err)
		return nil, err
	}

	// para cada carta, verifica se a carta tem valor que dobra
	roomData.DoubledCardValues = map[structs.TypePlayerPlays]structs.DoubledCardValues{
		structs.PlayPoliticalCard: {
			TimesValueDoubled:   0,
			RoundsUntilDecrease: 0,
		},
		structs.PlayRebelCard: {
			TimesValueDoubled:   0,
			RoundsUntilDecrease: 0,
		},
	}

	// preenche o deck com as cartas, repetindo de acordo com o número de cópias definido nas regras
	for copy := 0; copy < *generalRules.CopysOfEachCard; copy++ {
		roomData.Deck = append(roomData.Deck, structs.AllCards...)
	}

	// embaralha o deck
	utils.ShuffleSlice(roomData.Deck)

	// define o evento inicial da sala como "esperando início"
	roomData.GameEvent = structs.NewGameEvent("", structs.EventWaitingStart, time.Time{}, nil)

	// Salva a sala com TTL inicial de 5 segundos
	err = roomData.SaveRoomWithTTL(r.Context(), s.redisClient, time.Duration(s.roomTTL)*time.Second)
	if err != nil {
		return nil, err
	}

	return roomData, nil
}

func (s *RoomService) ValidatePlayerEntry(r *http.Request, rdb *redis.Client, backendService *BackendService, roomId string, playerId string) (bool, error) {
	_, err := backendService.GetToBackend("/getUserInfoById/" + playerId)
	if err != nil {
		return false, fmt.Errorf("Erro ao verificar player")
	}

	roomData, err := redisFuncs.LoadRoom(r.Context(), rdb, roomId)
	if err != nil {
		return false, fmt.Errorf("Sala não encontrada")
	}

	player, _ := roomData.GetPlayer(playerId)

	if player == nil && len(roomData.Players) >= roomData.MaxPlayers {
		return false, fmt.Errorf("Sala está cheia")
	}

	if !roomData.StartTime.IsZero() && player == nil {
		return false, fmt.Errorf("Jogo já começou")
	}

	if player != nil {
		_, isPlaying, err := playerRedis.GetRegisteredRoomForPlayer(r.Context(), rdb, playerId)

		if err != nil {
			return false, fmt.Errorf("Erro ao verificar status do player")
		}
		if isPlaying {
			return false, fmt.Errorf("Player já está em uma sala")
		}
	}

	return true, nil
}

func (s *RoomService) SyncActivePlayers(r *http.Request, rdb *redis.Client, roomData *rooms.Room) error {
	// Itera sobre os players e remove aqueles sem registro ativo
	for playerId := range roomData.Players {
		registeredRoom, registered, err := playerRedis.GetRegisteredRoomForPlayer(r.Context(), rdb, playerId)
		if err != nil {
			return err
		}
		if !registered || registeredRoom != roomData.ID {
			delete(roomData.Players, playerId)
		}
	}

	// Salva a sala atualizada
	return roomData.SaveRoom(r.Context(), rdb)
}
