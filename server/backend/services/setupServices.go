package services

import (
	"gorm.io/gorm"
)

type Services struct {
	UserService *UserService
	// RoomService    *RoomService
	// MessageService *MessageService
}

func SetupServices(db *gorm.DB) *Services {
	userService := NewUserService(db)
	// roomService := NewRoomService(db, userService)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		UserService: userService,
		// RoomService:    roomService,
		// MessageService: messageService,
	}
}
