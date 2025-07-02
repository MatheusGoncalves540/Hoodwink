package config

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// CheckEnvVars verifica se todas as variáveis de .env.example estão definidas no ambiente.
// Caso alguma esteja faltando, encerra o programa.
func CheckEnvVars(exampleFilePath string) {
	file, err := os.Open(exampleFilePath)
	if err != nil {
		log.Fatalf("Erro ao abrir %s: %v", exampleFilePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	missingVars := []string{}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])

		if os.Getenv(key) == "" {
			missingVars = append(missingVars, key)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro ao ler %s: %v", exampleFilePath, err)
	}

	if len(missingVars) > 0 {
		log.Fatalf("As seguintes variáveis de ambiente estão ausentes:\n  - %s", strings.Join(missingVars, "\n  - "))
	}
}
