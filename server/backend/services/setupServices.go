package services

import (
	"gorm.io/gorm"
)

type Services struct {
	UserService  *UserService
	NewDBService *DBService
	// RoomService    *RoomService
	// MessageService *MessageService
}

func SetupServices(db *gorm.DB) *Services {
	userService := NewUserService(db)
	newDBService := NewDBService(db)
	// roomService := NewRoomService(db, userService)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		UserService:  userService,
		NewDBService: newDBService,
		// RoomService:    roomService,
		// MessageService: messageService,
	}
}
