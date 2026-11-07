package services

import (
	"log"

	"gorm.io/gorm"
)

type Services struct {
	DB                *gorm.DB
	UserService       *UserService
	DBService         *DBService
	JWTService        *JWTService
	TicketService     *TicketService
	RequestLogService *RequestLogService
	ReplayService     *ReplayService
	// RoomService    *RoomService
	// MessageService *MessageService
}

func SetupServices(db *gorm.DB) *Services {
	userService := NewUserService(db)
	DBService := NewDBService(db)
	JWTService := NewJWTService()
	ticketService, err := NewTicketService()
	if err != nil {
		log.Fatalf("Erro ao carregar o serviço de ticket: %v", err)
	}
	requestLogService := NewRequestLogService(db)
	replayService, err := NewReplayService()
	if err != nil {
		log.Fatalf("Erro ao carregar a chave de assinatura de replay: %v", err)
	}
	// roomService := NewRoomService(db, userService)
	// messageService := NewMessageService(db, userService, roomService)

	return &Services{
		DB:                db,
		UserService:       userService,
		DBService:         DBService,
		JWTService:        JWTService,
		TicketService:     ticketService,
		RequestLogService: requestLogService,
		ReplayService:     replayService,
		// RoomService:    roomService,
		// MessageService: messageService,
	}
}
