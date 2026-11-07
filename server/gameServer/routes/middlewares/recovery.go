package middlewares

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/routes/contextKeys"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/services"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// maxBodyCapture limita a leitura do corpo para o coletor (64 KB, plano 04).
const maxBodyCapture = 64 * 1024

// reqInfoContextKey guarda o ponteiro *ReqInfo preenchido pelo JWTBackendMiddleware.
type reqInfoContextKeyType string

const reqInfoContextKey reqInfoContextKeyType = "reqInfo"

// ReqInfo é preenchido pelo middleware de auth e lido pelo RequestMiddleware.
type ReqInfo struct {
	UserID   string
	Username string
}

// APIResponse é uma estrutura de resposta padronizada
type APIResponse struct {
	Error   any    `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// RequestMiddleware faz log/recovery/trace ID e envia a request ao coletor.
// Não registra /alive nem /internal/*.
func RequestMiddleware(next http.Handler, shipper *services.RequestLogShipper) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()
		reqID, err := uuid.NewV7()
		if err != nil {
			utils.LogError(err.Error())
		}

		info := &ReqInfo{}
		ctx := context.WithValue(r.Context(), contextKeys.RequestIDKey, reqID)
		ctx = context.WithValue(ctx, reqInfoContextKey, info)
		r = r.WithContext(ctx)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		skip := shouldSkipRequestLog(r.URL.Path) || shipper == nil

		var bodyText string
		if !skip {
			bodyText = captureBody(r)
		}

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

			if skip {
				return
			}
			shipper.Enqueue(buildRequestLogEntry(r, info, reqID, rw.statusCode, duration, bodyText))
		}()

		next.ServeHTTP(rw, r)
	})
}

func shouldSkipRequestLog(path string) bool {
	if path == "/alive" {
		return true
	}
	return strings.HasPrefix(path, "/internal/")
}

func captureBody(r *http.Request) string {
	if r.Body == nil || r.Body == http.NoBody {
		return ""
	}

	// +1 byte detecta corpo acima de 64 KB (grava só <N bytes>).
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBodyCapture+1))
	if err != nil {
		data = data[:0]
	}
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(data), r.Body))

	if len(data) == 0 {
		return ""
	}
	if len(data) > maxBodyCapture {
		return byteCount(len(data))
	}
	return utils.MaskBody(r.Header.Get("Content-Type"), string(data))
}

func byteCount(n int) string {
	return fmt.Sprintf("<%d bytes>", n)
}

func buildRequestLogEntry(r *http.Request, info *ReqInfo, reqID uuid.UUID, status int, duration time.Duration, bodyText string) *services.RequestLogEntry {
	method := r.Method
	path := r.URL.Path
	ip := utils.ClientIP(r)
	ua := utils.UserAgent(r)
	requestID := reqID.String()
	durationMs := int(duration.Milliseconds())

	entry := &services.RequestLogEntry{
		At:         time.Now().UTC(),
		Service:    "gameServer",
		Method:     &method,
		Path:       &path,
		Status:     &status,
		DurationMs: &durationMs,
		IP:         &ip,
		UserAgent:  &ua,
		RequestID:  &requestID,
	}

	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if route := rctx.RoutePattern(); route != "" {
			entry.Route = &route
		}
	}
	if info.UserID != "" {
		uid := info.UserID
		entry.UserID = &uid
	}
	if info.Username != "" {
		username := info.Username
		entry.Username = &username
	}
	if bodyText != "" {
		entry.Body = &bodyText
	}

	return entry
}

// responseWriter captura status code e mantém Hijack/Flush para o WebSocket.
type responseWriter struct {
	ResponseWriter http.ResponseWriter
	statusCode     int
	wroteHeader    bool
}

func (rw *responseWriter) Header() http.Header {
	return rw.ResponseWriter.Header()
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.wroteHeader = true
	}
}

// Flush implementa http.Flusher para streaming (ex.: upgrade/SSE).
func (rw *responseWriter) Flush() {
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implementa http.Hijacker para suportar WebSocket.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
	}
	return hj.Hijack()
}
