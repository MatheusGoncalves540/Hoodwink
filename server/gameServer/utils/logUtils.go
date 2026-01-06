package utils

import (
	"fmt"
	"os"
)

// if DEBUG = true, print on console
func LogDebug(msg any) {
	if os.Getenv("DEBUG") == "true" {
		fmt.Println(msg)
	}
}

// LogError sempre usado para printar mensagem de erro no console
func LogError(msg any) {
	fmt.Println(msg)
}

func LogInvldPlyrReq(msg any) {
	fmt.Println(msg)
}
