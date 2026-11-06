package utils

import (
	"log"
	"os"
)

// if DEBUG = true, print on console
func LogDebug(msg any) {
	if os.Getenv("DEBUG") == "true" {
		log.Println(msg)
	}
}

// LogError sempre usado para printar mensagem de erro no console
func LogError(msg any) {
	log.Println(msg)
}

func LogInvldPlyrReq(msg any, playerId string) {
	log.Println(msg)
}
