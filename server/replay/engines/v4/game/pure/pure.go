// Package pure reúne as utilidades sem dependência de rede, env ou validator.
// O fecho do motor (game/engine, game/structs, game/structs/rooms, game/rules e
// handlers/websocket/messages) usa este pacote no lugar do pacote utils, que puxa
// go-redis/gorilla/validator e não pode entrar no binário de replay (WASM).
package pure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// DecodeStrictJSON decodifica JSON em uma struct, retornando erro se houver campos desconhecidos.
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

// LogDebug imprime no console apenas quando DEBUG=true.
func LogDebug(msg any) {
	if os.Getenv("DEBUG") == "true" {
		log.Println(msg)
	}
}

// LogError sempre imprime a mensagem de erro no console.
func LogError(msg any) {
	log.Println(msg)
}

// LogInvldPlyrReq imprime a mensagem de requisição inválida do jogador no console.
func LogInvldPlyrReq(msg any, playerId string) {
	log.Println(msg)
}

// FloorDiv divide a por b arredondando o resultado para baixo (piso matemático).
// Todos os cálculos de taxas do jogo usam arredondamento para baixo.
func FloorDiv(a, b int) int {
	if b == 0 {
		return 0
	}

	q := a / b
	// Em Go, o resto tem o sinal do dividendo. Se o resto não for zero e os
	// sinais forem diferentes, o quociente truncado ficou acima do piso.
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

// CeilDiv divide a por b arredondando o resultado para cima (teto matemático).
func CeilDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return -FloorDiv(-a, b)
}
