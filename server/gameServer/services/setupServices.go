package services

import "github.com/redis/go-redis/v9"

type Services struct {
	RoomService      *RoomService
	JWTService       *JWTService
	BackendService   *BackendService
	HeartbeatService *HeartbeatService
	// MessageService *MessageService
}

func SetupServices(redisClient *redis.Client) *Services {
	roomService := NewRoomService(redisClient)
	JWTService := NewJWTService()
	backendService := NewBackendService()
	heartbeatService := NewHeartbeatService(redisClient)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		RoomService:      roomService,
		JWTService:       JWTService,
		BackendService:   backendService,
		HeartbeatService: heartbeatService,
		// MessageService: messageService,
	}
}
