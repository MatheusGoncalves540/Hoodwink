package services

import (
	"sync"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
)

// capturaLotes cria um serviço de teste que desvia a gravação para um coletor.
func capturaLotes(t *testing.T) (*RequestLogService, *[]*models.RequestLog, *sync.Mutex) {
	t.Helper()
	var mu sync.Mutex
	var got []*models.RequestLog
	s := NewRequestLogServiceForTest(func(batch []*models.RequestLog) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, batch...)
	})
	t.Cleanup(s.Close)
	return s, &got, &mu
}

func TestRequestLogServiceCloseDrena(t *testing.T) {
	s, got, mu := capturaLotes(t)

	s.Enqueue(&models.RequestLog{ID: 1, Service: "backend"})
	s.Enqueue(&models.RequestLog{ID: 2, Service: "backend"})

	s.Close()

	mu.Lock()
	defer mu.Unlock()
	if len(*got) != 2 {
		t.Fatalf("Close() deveria drenar 2 linhas, gravou %d", len(*got))
	}
}

func TestRequestLogServiceDescartaQuandoCheio(t *testing.T) {
	blockCh := make(chan struct{})
	s := NewRequestLogServiceForTest(func([]*models.RequestLog) { <-blockCh })

	// Mais do que (buffer + 1 lote) garante descarte mesmo com o loop consumindo.
	for i := 0; i < requestLogChannelSize+200; i++ {
		s.Enqueue(&models.RequestLog{ID: int64(i + 1), Service: "backend"})
	}

	if got := s.dropped.Load(); got == 0 {
		t.Fatalf("esperado algum descarte com canal cheio, veio 0")
	}

	close(blockCh) // destrava o loop para permitir Close
	s.Close()
}

func TestRequestLogServiceEnqueueNilIgnora(t *testing.T) {
	s := NewRequestLogServiceForTest(func([]*models.RequestLog) {})
	s.Enqueue(nil)
	if got := s.dropped.Load(); got != 0 {
		t.Fatalf("nil não deve descartar, veio %d", got)
	}
	s.Close()
}
