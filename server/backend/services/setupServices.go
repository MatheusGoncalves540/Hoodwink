package services

import (
	"gorm.io/gorm"
)

type Services struct {
	UserService *UserService
	DBService   *DBService
	// RoomService    *RoomService
	// MessageService *MessageService
}

func SetupServices(db *gorm.DB) *Services {
	userService := NewUserService(db)
	DBService := NewDBService(db)
	// roomService := NewRoomService(db, userService)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		UserService: userService,
		DBService:   DBService,
		// RoomService:    roomService,
		// MessageService: messageService,
	}
}
