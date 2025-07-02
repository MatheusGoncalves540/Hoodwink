package services

import "github.com/redis/go-redis/v9"

type Services struct {
	RoomService *RoomService
	// MessageService *MessageService
}

func SetupServices(redisClient *redis.Client) *Services {
	roomService := NewRoomService(redisClient)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		RoomService: roomService,
		// MessageService: messageService,
	}
}
