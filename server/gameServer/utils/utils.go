package utils

import (
	"bytes"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateInfos valida a struct e retorna false se houver erros (já escreve resposta JSON)
func ValidateInfos(w http.ResponseWriter, toValidate any) bool {
	if err := validate.Struct(toValidate); err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages,
				fmt.Sprintf("Campo '%s' inválido: %s", err.Field(), err.Tag()))
		}

		SendJSON(w, http.StatusBadRequest, APIResponse{
			Error:   errorMessages,
			Message: "Erro de validação nos dados enviados",
		})
		return false
	}

	return true
}

// roomCodeAlphabet é o alfabeto do código da sala: dígitos e maiúsculas sem os
// que se confundem na leitura (0/O, 1/I/L). 32 caracteres para o índice sair de
// um byte sem viés (256 é múltiplo de 32).
const roomCodeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

// roomCodeLength é o tamanho do código que o jogador digita e da URL `?sala=`.
const roomCodeLength = 6

// GenerateNewId cria o código da sala: roomCodeLength caracteres do
// roomCodeAlphabet com entropia criptográfica.
func GenerateNewId() string {
	code := make([]byte, roomCodeLength)
	if _, err := crand.Read(code); err != nil {
		// Entropia indisponível (não ocorre em plataformas suportadas): usa o
		// relógio como último recurso para nunca deixar a sala sem id.
		n := uint64(time.Now().UTC().UnixNano())
		for i := range code {
			code[i] = roomCodeAlphabet[n&31]
			n >>= 5
		}
		return string(code)
	}
	for i := range code {
		code[i] = roomCodeAlphabet[code[i]&31]
	}
	return string(code)
}

// GetRandomElementFromSlice retorna um elemento aleatório de um slice genérico
func GetRandomElementFromSlice[T any](slice []T) T {
	randomIndex := rand.IntN(len(slice))
	return slice[randomIndex]
}

// DecodeStrictJSON decodifica JSON em uma struct, retornando erro se houver campos desconhecidos
func DecodeStrictJSON(data []byte, target any) error {
	if target == nil {
		return fmt.Errorf("target não pode ser nil")
	}

	reader := bytes.NewReader(data)
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("json inválido: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("json contém dados extras não esperados")
	}

	return nil
}

// ShuffleSlice embaralha um slice genérico
func ShuffleSlice[T any](slice []T) {
	rand.Shuffle(len(slice), func(i, j int) {
		slice[i], slice[j] = slice[j], slice[i]
	})
}
