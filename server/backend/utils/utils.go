package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateInfos valida a struct e retorna false se houver erros (já escreve resposta JSON)
func ValidateInfos(w http.ResponseWriter, toValidate any) bool {
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

func LogError(msg any) {
	log.Println(msg)
}

// if DEBUG = true, print on console
func LogDebug(msg any) {
	if os.Getenv("DEBUG") == "true" {
		log.Println(msg)
	}
}
