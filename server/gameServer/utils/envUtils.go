package utils

import (
	"os"
	"strconv"
)

// IsDebugMode retorna true se o modo debug estiver ativado via variável de ambiente.
func IsDebugMode() bool {
	return os.Getenv("DEBUG") == "true"
}

// MustEnvInt lê uma variável de ambiente e converte para int, ou retorna valor padrão se não definida ou inválida.
func MustEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
