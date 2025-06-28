package routesFuncs

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type CreateRoomRequest struct {
	RoomName string  `json:"roomName" validate:"required,max=30,min=3"`
	Password *string `json:"password" validate:"omitempty,max=24"`
}

var validate = validator.New()

// Create a new room with the given data
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	var reqBody CreateRoomRequest

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Validação com o validator
	if err := validate.Struct(reqBody); err != nil {
		// Formatar os erros de validação
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Campo '%s' inválido: %s", err.Field(), err.Tag()))
		}

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": errorMessages,
		})
		return
	}

	// Aqui está tudo validado
	msg := fmt.Sprintf("Criou uma sala chamada %s", reqBody.RoomName)
	if reqBody.Password != nil {
		msg += " com senha definida."
	} else {
		msg += " sem senha."
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}
