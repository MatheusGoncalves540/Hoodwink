package routesFuncs

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Create a new room with the given data
func CreateRoom(w http.ResponseWriter, r *http.Request) {
	var roomName string = "teste"

	frase := fmt.Sprintf("Criou uma sala chamada %s", roomName)

	json.NewEncoder(w).Encode(map[string]string{"message": frase})
}
