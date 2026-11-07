package services

import (
	"errors"
	"testing"
)

func TestRunRetentionLoopMoveUmaVezSo(t *testing.T) {
	calls := 0
	err := runRetentionLoop(func() (int64, error) {
		calls++
		return 0, nil // sem linhas: move uma vez e para
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if calls != 1 {
		t.Fatalf("sem linhas deveria chamar move 1 vez, chamou %d", calls)
	}
}

func TestRunRetentionLoopMoveEmLotesAteEsgotar(t *testing.T) {
	remaining := []int64{10000, 10000, 5000}
	calls := 0
	err := runRetentionLoop(func() (int64, error) {
		if calls >= len(remaining) {
			return 0, nil
		}
		n := remaining[calls]
		calls++
		return n, nil
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if calls != 3 {
		t.Fatalf("esperado 3 lotes (10000,10000,5000), chamou %d", calls)
	}
}

func TestRunRetentionLoopPropagaErro(t *testing.T) {
	want := errors.New("banco fora")
	err := runRetentionLoop(func() (int64, error) { return 0, want })
	if !errors.Is(err, want) {
		t.Fatalf("esperado erro propagado %v, veio %v", want, err)
	}
}
