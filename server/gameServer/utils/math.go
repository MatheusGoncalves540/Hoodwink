package utils

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
