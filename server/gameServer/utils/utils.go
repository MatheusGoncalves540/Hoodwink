package utils

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
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

// GenerateNewId cria um ID hex de 16 caracteres (8 bytes) com entropia
// criptográfica. O formato (hex minúsculo) é preservado para os callers.
func GenerateNewId() string {
	b := make([]byte, 8)
	if _, err := crand.Read(b); err != nil {
		// Entropia indisponível (não ocorre em plataformas suportadas): usa o
		// relógio como último recurso para nunca deixar a sala sem id.
		return strconv.FormatInt(time.Now().UTC().UnixMilli(), 16)
	}
	return hex.EncodeToString(b)
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
