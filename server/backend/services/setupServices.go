package services

import (
	"gorm.io/gorm"
)

type Services struct {
	UserService *UserService
	DBService   *DBService
	JWTService  *JWTService
	// RoomService    *RoomService
	// MessageService *MessageService
}

func SetupServices(db *gorm.DB) *Services {
	userService := NewUserService(db)
	DBService := NewDBService(db)
	JWTService := NewJWTService()
	// roomService := NewRoomService(db, userService)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		UserService: userService,
		DBService:   DBService,
		JWTService:  JWTService,
		// RoomService:    roomService,
		// MessageService: messageService,
	}
}
