package utils

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// SendJSON envia um JSON com o status HTTP e payload genérico
func SendJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// SendSuccess envia uma resposta JSON de sucesso com dados opcionais
func SendSuccess(w http.ResponseWriter, data any, message string) {
	SendJSON(w, http.StatusOK, APIResponse{
		Data:    data,
		Message: message,
	})
}

// SendError envia uma resposta JSON de erro com status HTTP e mensagem
func SendError(w http.ResponseWriter, errMessage string, status int) {
	SendJSON(w, status, APIResponse{
		Error:   errMessage,
		Message: "Erro ao processar a requisição",
	})
}
