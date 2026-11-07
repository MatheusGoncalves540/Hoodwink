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
	"github.com/google/uuid"
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

	matchID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("gerar matchId: %w", err)
	}
	seed := rooms.NewSeed()

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
		MatchID:                   matchID.String(),
		Seed:                      seed,
		Rng:                       rooms.NewRngState(seed),
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

	// embaralha o deck com o RNG semeado da sala (determinístico para o replay)
	roomData.Rand().Shuffle(len(roomData.Deck), func(i, j int) {
		roomData.Deck[i], roomData.Deck[j] = roomData.Deck[j], roomData.Deck[i]
	})

	// define o evento inicial da sala como "esperando início"
	roomData.GameEvent = structs.NewGameEvent("", structs.EventWaitingStart, time.Time{}, nil)

	// Salva a sala com TTL inicial de 5 segundos
	err = redisFuncs.SaveRoomWithTTL(r.Context(), s.redisClient, roomData, time.Duration(s.roomTTL)*time.Second)
	if err != nil {
		return nil, err
	}

	return roomData, nil
}

func (s *RoomService) ValidatePlayerEntry(r *http.Request, rdb *redis.Client, roomId string, playerId string) (bool, error) {
	roomData, err := redisFuncs.LoadRoom(r.Context(), rdb, roomId)
	if err != nil {
		return false, fmt.Errorf("Sala não encontrada")
	}

	player, _ := roomData.GetPlayer(playerId)

	if player == nil && roomData.IsFull() {
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

func (s *RoomService) SyncActivePlayers(r *http.Request, rdb *redis.Client, roomData *rooms.Room) (*rooms.Room, error) {
	// Remove players sem registro ativo (desconectados) quando o jogo ainda não começou.
	fresh := roomData
	err := redisFuncs.WithRoom(r.Context(), rdb, roomData.ID, redisFuncs.CauseLeave, func(room *rooms.Room) bool {
		fresh = room
		aplicado := false
		for playerId := range room.Players {
			if !room.IsConnected(playerId) {
				delete(room.Players, playerId)
				room.RemoveConnected(playerId)
				aplicado = true
			}
		}
		return aplicado
	})
	if err != nil {
		return nil, err
	}
	return fresh, nil
}
