package services

import (
	"gorm.io/gorm"
)

type Services struct {
	RoomService *RoomService
	// MessageService *MessageService
}

func SetupServices(db *gorm.DB) *Services {
	roomService := NewRoomService(db)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		RoomService: roomService,
		// MessageService: messageService,
	}
}
