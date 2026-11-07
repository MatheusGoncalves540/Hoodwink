//go:build !js || !wasm

package main

import "fmt"

// Fora de js/wasm o binário não faz nada: existe só para `go build ./...`,
// `go vet ./...` e `go test ./...` continuarem funcionando no host. O build de
// verdade é feito com GOOS=js GOARCH=wasm (make replay-wasm).
func main() {
	fmt.Println("cmd/wasm: compile com GOOS=js GOARCH=wasm (make replay-wasm)")
}
