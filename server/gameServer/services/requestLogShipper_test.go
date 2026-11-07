package services

import (
	"errors"
	"sync"
	"testing"
)

func TestRequestLogShipperCloseDrena(t *testing.T) {
	var mu sync.Mutex
	var got []*RequestLogEntry
	s := NewRequestLogShipperForTest(func(batch []*RequestLogEntry) error {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, batch...)
		return nil
	})

	s.Enqueue(&RequestLogEntry{Service: "gameServer"})
	s.Enqueue(&RequestLogEntry{Service: "gameServer"})

	s.Close()

	mu.Lock()
	defer mu.Unlock()
	if len(got) != 2 {
		t.Fatalf("Close() deveria drenar 2 linhas, enviou %d", len(got))
	}
}

func TestRequestLogShipperBackendErroDescarta(t *testing.T) {
	calls := 0
	s := NewRequestLogShipperForTest(func(batch []*RequestLogEntry) error {
		calls++
		return errors.New("backend fora")
	})

	s.Enqueue(&RequestLogEntry{Service: "gameServer"})
	s.Close()

	if calls != 1 {
		t.Fatalf("flush deveria ser chamado 1 vez (descarta sem retry), chamou %d", calls)
	}
}

func TestRequestLogShipperEnqueueNilIgnora(t *testing.T) {
	s := NewRequestLogShipperForTest(func([]*RequestLogEntry) error { return nil })
	s.Enqueue(nil)
	if got := s.dropped.Load(); got != 0 {
		t.Fatalf("nil não deve descartar, veio %d", got)
	}
	s.Close()
}
