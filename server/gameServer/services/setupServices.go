package services

import (
	"github.com/redis/go-redis/v9"
)

type Services struct {
	RoomService       *RoomService
	JWTService        *JWTService
	BackendService    *BackendService
	HeartbeatService  *HeartbeatService
	RequestLogShipper *RequestLogShipper
	// MessageService *MessageService
}

func SetupServices(redisClient *redis.Client) *Services {
	roomService := NewRoomService(redisClient)
	JWTService := NewJWTService()
	backendService := NewBackendService()
	heartbeatService := NewHeartbeatService(redisClient)
	requestLogShipper := NewRequestLogShipper()
	// messageService := NewMessageService(db, userService, roomService)

	services := &Services{
		RoomService:       roomService,
		JWTService:        JWTService,
		BackendService:    backendService,
		HeartbeatService:  heartbeatService,
		RequestLogShipper: requestLogShipper,
		// MessageService: messageService,
	}

	InitReplayRecorder(backendService)

	return services
}
