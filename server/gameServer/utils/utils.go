package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/room/roomStructs"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var validate = validator.New()

// ValidateInfos valida a struct e retorna false se houver erros (já escreve resposta JSON)
func ValidateInfos(w http.ResponseWriter, toValidate interface{}) bool {
	if err := validate.Struct(toValidate); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages,
				fmt.Sprintf("Campo '%s' inválido: %s", err.Field(), err.Tag()))
		}

		SendJSON(w, http.StatusBadRequest, APIResponse{
			Error:   errorMessages,
			Message: "Erro de validação nos dados enviados",
		})
		return false
	}

	return true
}

// GenerateNewId cria um ID de 16 caracteres baseado em timestamp e fator randômico
func GenerateNewId() string {
	rand.NewSource(time.Now().UnixNano())
	randomFactor := rand.Float64() * rand.Float64()
	id := int64(randomFactor * float64(time.Now().UnixNano()))
	return strconv.FormatInt(id, 16)
}

// Gera um UUID para identificar a instância/processo
func GetInstanceID() string {
	return uuid.New().String()
}

// StringContains verifica se uma string está presente em um slice de strings.
func StringContains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

// GetPlayerByID retorna o ponteiro para o jogador com o ID fornecido, ou nil se não encontrado.
func GetPlayerByID(room *roomStructs.Room, playerID string) *roomStructs.Player {
	for i := range room.Players {
		if room.Players[i].Id == playerID {
			return &room.Players[i]
		}
	}
	return nil
}

// if DEBUG = true, print on console
func LogDebug(msg any) {
	if os.Getenv("DEBUG") == "true" {
		fmt.Println(msg)
	}
}
