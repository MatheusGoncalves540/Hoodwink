package utils

import (
	"encoding/json"
	"io"
	"mime"
	"strconv"
	"strings"
	"unicode/utf8"
)

// maxMaskedBody controla o corte final do corpo gravado (plano 04).
const maxMaskedBody = 4096

var sensitiveBodyKeys = []string{
	"token",
	"ticket",
	"password",
	"senha",
	"secret",
	"authorization",
	"apikey",
	"api-key",
	"credential",
	"jwt",
}

// MaskBody devolve o corpo pronto para gravar em request_logs: JSON tem as
// chaves sensíveis mascaradas em qualquer profundidade e é cortado em 4096
// bytes. Corpo vazio continua vazio; JSON inválido ou não-JSON vira <N bytes>.
func MaskBody(contentType, raw string) string {
	if raw == "" {
		return ""
	}
	if !isJSONContentType(contentType) {
		return byteCount(raw)
	}

	var body any
	if !decodeJSONBody(raw, &body) {
		return byteCount(raw)
	}

	masked, err := json.Marshal(maskJSON(body))
	if err != nil {
		return byteCount(raw)
	}

	return truncateUTF8(string(masked), maxMaskedBody)
}

func decodeJSONBody(raw string, dst *any) bool {
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()

	if err := decoder.Decode(dst); err != nil {
		return false
	}

	// Rejeita lixo depois do primeiro valor (ex.: {"a":1} {"b":2}).
	if err := decoder.Decode(new(any)); err != io.EOF {
		return false
	}

	return true
}

func maskJSON(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if isSensitiveBodyKey(key) {
				typed[key] = "***"
				continue
			}
			typed[key] = maskJSON(item)
		}
		return typed
	case []any:
		for i, item := range typed {
			typed[i] = maskJSON(item)
		}
		return typed
	default:
		return value
	}
}

func isSensitiveBodyKey(key string) bool {
	lowerKey := strings.ToLower(key)
	for _, sensitive := range sensitiveBodyKeys {
		if strings.Contains(lowerKey, sensitive) {
			return true
		}
	}
	return false
}

func isJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		mediaType = strings.TrimSpace(strings.SplitN(contentType, ";", 2)[0])
	}

	mediaType = strings.ToLower(mediaType)
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func truncateUTF8(value string, limit int) string {
	if len(value) <= limit {
		return value
	}

	value = value[:limit]
	for !utf8.ValidString(value) {
		value = value[:len(value)-1]
	}
	return value
}

func byteCount(raw string) string {
	return "<" + strconv.Itoa(len(raw)) + " bytes>"
}
