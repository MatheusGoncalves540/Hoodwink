package middlewares

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ContextKey é uma chave customizada para armazenar dados no contexto
type ContextKey string

const RequestIDKey ContextKey = "requestID"

// reqInfoContextKey guarda o ponteiro *ReqInfo preenchido pelos middlewares de
// auth (JWT) e lido pelo RequestMiddleware depois do next.
type reqInfoContextKeyType string

const reqInfoContextKey reqInfoContextKeyType = "reqInfo"

// maxBodyCapture limita a leitura do corpo para o coletor (64 KB, plano 04).
const maxBodyCapture = 64 * 1024

// ReqInfo é preenchido pelo middleware de auth e lido pelo RequestMiddleware
// para gravar user_id e username sem consultar o banco.
type ReqInfo struct {
	UserID   string
	Username string
}

// APIResponse é uma estrutura de resposta padronizada
type APIResponse struct {
	Error   any    `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// RequestMiddleware faz log/recovery/trace ID e grava a request no coletor.
// Não grava /alive nem /internal/*; /bugReports* grava só o tamanho do corpo.
func RequestMiddleware(next http.Handler, logSvc *services.RequestLogService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().UTC()
		reqID, err := uuid.NewV7()
		if err != nil {
			utils.LogError(err.Error())
		}

		info := &ReqInfo{}
		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		ctx = context.WithValue(ctx, reqInfoContextKey, info)
		r = r.WithContext(ctx)

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		skip := shouldSkipRequestLog(r.URL.Path) || logSvc == nil

		var bodyText string
		if !skip && shouldCaptureBody(r) {
			bodyText = captureBody(r)
		}

		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[ERRO] [%s] panic: %v", reqID, rec)
				utils.SendError(rw, "Ocorreu um erro no servidor. Tente novamente mais tarde.", http.StatusInternalServerError)
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
			logSvc.Enqueue(buildRequestLog(r, info, reqID, rw.statusCode, duration, bodyText))
		}()

		next.ServeHTTP(rw, r)
	})
}

// shouldSkipRequestLog decide se a request é ruído/loop e não deve ser gravada.
func shouldSkipRequestLog(path string) bool {
	if path == "/alive" {
		return true
	}
	return strings.HasPrefix(path, "/internal/")
}

// shouldCaptureBody diz se vale ler o corpo para o coletor: só métodos que levam
// body (POST/PUT/PATCH) e com Content-Length > 0. GET/HEAD/DELETE e corpo chunked
// (Content-Length -1) não pagam o custo da leitura.
func shouldCaptureBody(r *http.Request) bool {
	if r.ContentLength <= 0 {
		return false
	}
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		return true
	default:
		return false
	}
}

// captureBody lê no máximo 64 KB do corpo e devolve ao handler com MultiReader,
// então devolve a string pronta para gravar (mascarada/cortada em 4 KB).
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

	// /bugReports* grava só o tamanho do corpo (plano 06).
	if strings.HasPrefix(r.URL.Path, "/bugReports") {
		return byteCount(len(data))
	}

	if len(data) > maxBodyCapture {
		return byteCount(len(data))
	}
	return utils.MaskBody(r.Header.Get("Content-Type"), string(data))
}

func byteCount(n int) string {
	return fmt.Sprintf("<%d bytes>", n)
}

func buildRequestLog(r *http.Request, info *ReqInfo, reqID uuid.UUID, status int, duration time.Duration, bodyText string) *models.RequestLog {
	now := time.Now().UTC()
	method := r.Method
	path := r.URL.Path
	ip := utils.ClientIP(r)
	ua := utils.UserAgent(r)
	requestID := reqID.String()
	durationMs := int(duration.Milliseconds())

	entry := &models.RequestLog{
		At:         now,
		Service:    "backend",
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
