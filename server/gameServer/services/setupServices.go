package services

import "github.com/redis/go-redis/v9"

type Services struct {
	RoomService    *RoomService
	JWTService     *JWTService
	BackendService *BackendService
	// MessageService *MessageService
}

func SetupServices(redisClient *redis.Client) *Services {
	roomService := NewRoomService(redisClient)
	JWTService := NewJWTService()
	backendService := NewBackendService()
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		RoomService:    roomService,
		JWTService:     JWTService,
		BackendService: backendService,
		// MessageService: messageService,
	}
}
