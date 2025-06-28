package middlewares

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ContextKey é uma chave customizada para armazenar dados no contexto
type ContextKey string

const RequestIDKey ContextKey = "requestID"

// APIResponse é uma estrutura de resposta padronizada
type APIResponse struct {
	Error   interface{} `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Middleware para log, recovery e trace ID
func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := uuid.New().String()

		// Anexa request ID ao contexto
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		r = r.WithContext(ctx)

		// Cria um ResponseWriter customizado para capturar status
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[ERRO] [%s] panic: %v", reqID, rec)
				rw.Header().Set("Content-Type", "application/json")
				rw.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(rw).Encode(APIResponse{
					Error:   "Erro interno inesperado",
					Message: "Ocorreu um erro no servidor. Tente novamente mais tarde.",
				})
			}

			duration := time.Since(start)
			log.Printf("[INFO] [%s] %s %s %d %s",
				reqID,
				r.Method,
				r.URL.Path,
				rw.statusCode,
				duration,
			)
		}()

		next.ServeHTTP(rw, r)
	})
}

// responseWriter captura status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// Sobrescreve WriteHeader para registrar status
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
