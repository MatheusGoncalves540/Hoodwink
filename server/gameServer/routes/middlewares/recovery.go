package middlewares

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/google/uuid"
)

// APIResponse é uma estrutura de resposta padronizada
type APIResponse struct {
	Error   any    `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// Middleware para log, recovery e trace ID
func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID, err := uuid.NewV7()
		if err != nil {
			utils.LogError(err.Error())
		}

		// Anexa request ID ao contexto
		ctx := context.WithValue(r.Context(), contextKeys.RequestIDKey, reqID)
		r = r.WithContext(ctx)

		// Cria um ResponseWriter customizado para capturar status
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[ERRO] [%s] panic: %v", reqID, rec)
				utils.SendError(rw, "Ocorreu um erro em nossos servidores :c. Tente novamente mais tarde.", http.StatusInternalServerError)
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
	ResponseWriter http.ResponseWriter
	statusCode     int
	wroteHeader    bool
}

// Implementa Header para compatibilidade com http.ResponseWriter
func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

// Implementa Write para compatibilidade com http.ResponseWriter
func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// Sobrescreve WriteHeader para registrar status
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.wroteHeader = true
	}
}

// Implementa http.Hijacker para suportar WebSocket
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
	}
	return hj.Hijack()
}
