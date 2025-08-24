package services

import "github.com/redis/go-redis/v9"

type Services struct {
	RoomService *RoomService
	JWTService  *JWTService
	// MessageService *MessageService
}

func SetupServices(redisClient *redis.Client) *Services {
	roomService := NewRoomService(redisClient)
	JWTService := NewJWTService()
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		RoomService: roomService,
		JWTService:  JWTService,
		// MessageService: messageService,
	}
}
