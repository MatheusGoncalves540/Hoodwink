package utils

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

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
