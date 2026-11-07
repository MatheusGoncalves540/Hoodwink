package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	shipperChannelSize     = 1000
	shipperBatchSize       = 100
	shipperFlushInterval   = 2 * time.Second
	shipperDropLogInterval = time.Minute
	shipperTimeout         = 3 * time.Second
)

// RequestLogEntry é uma linha de request do gameServer enviada ao backend.
type RequestLogEntry struct {
	At         time.Time `json:"at"`
	Service    string    `json:"service"`
	Method     *string   `json:"method,omitempty"`
	Path       *string   `json:"path,omitempty"`
	Route      *string   `json:"route,omitempty"`
	Status     *int      `json:"status,omitempty"`
	DurationMs *int      `json:"duration_ms,omitempty"`
	IP         *string   `json:"ip,omitempty"`
	UserAgent  *string   `json:"user_agent,omitempty"`
	UserID     *string   `json:"user_id,omitempty"`
	Username   *string   `json:"username,omitempty"`
	RequestID  *string   `json:"request_id,omitempty"`
	Body       *string   `json:"body,omitempty"`
	Error      *string   `json:"error,omitempty"`
}

// RequestLogShipper junta as linhas em lote e envia a cada 2s (ou 100 linhas)
// para POST /internal/requestLogs. Backend fora → descarta o lote e loga;
// canal cheio descarta (com contagem logada a cada minuto). Close drena no SIGTERM.
type RequestLogShipper struct {
	apiKey    string
	url       string
	client    *http.Client
	ch        chan *RequestLogEntry
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
	dropped   atomic.Int64
	// postFn é substituível em testes para observar os lotes sem backend.
	postFn func([]*RequestLogEntry) error
}

func NewRequestLogShipper() *RequestLogShipper {
	s := newRequestLogShipper()
	s.postFn = s.post
	s.start()
	return s
}

// NewRequestLogShipperForTest desvia o envio para fn (sem backend).
func NewRequestLogShipperForTest(fn func([]*RequestLogEntry) error) *RequestLogShipper {
	s := newRequestLogShipper()
	s.postFn = fn
	s.start()
	return s
}

func newRequestLogShipper() *RequestLogShipper {
	return &RequestLogShipper{
		apiKey: os.Getenv("BACKEND_API_KEY"),
		url:    os.Getenv("BACKEND_URL"),
		client: &http.Client{Timeout: shipperTimeout},
		ch:     make(chan *RequestLogEntry, shipperChannelSize),
		done:   make(chan struct{}),
	}
}

func (s *RequestLogShipper) start() {
	s.wg.Add(1)
	go s.loop()
}

// Enqueue enfileira uma linha; canal cheio descarta (contado pelo loop).
func (s *RequestLogShipper) Enqueue(entry *RequestLogEntry) {
	if entry == nil {
		return
	}
	select {
	case s.ch <- entry:
	default:
		s.dropped.Add(1)
	}
}

// Close drena o canal e envia o último lote; idempotente.
func (s *RequestLogShipper) Close() {
	s.closeOnce.Do(func() {
		close(s.done)
		s.wg.Wait()
	})
}

func (s *RequestLogShipper) loop() {
	defer s.wg.Done()

	batch := make([]*RequestLogEntry, 0, shipperBatchSize)
	ticker := time.NewTicker(shipperFlushInterval)
	dropTicker := time.NewTicker(shipperDropLogInterval)
	defer ticker.Stop()
	defer dropTicker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		s.flush(batch)
		batch = batch[:0]
	}

	for {
		select {
		case entry := <-s.ch:
			batch = append(batch, entry)
			if len(batch) >= shipperBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		case <-dropTicker.C:
			if n := s.dropped.Swap(0); n > 0 {
				log.Printf("requestlog dropped=%d", n)
			}
		case <-s.done:
			flush()
			for {
				select {
				case entry := <-s.ch:
					batch = append(batch, entry)
					if len(batch) >= shipperBatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}
		}
	}
}

func (s *RequestLogShipper) flush(batch []*RequestLogEntry) {
	if err := s.postFn(batch); err != nil {
		log.Printf("requestlog: descartando lote de %d linhas: %v", len(batch), err)
	}
}

func (s *RequestLogShipper) post(batch []*RequestLogEntry) error {
	body, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(s.url, "/")+"/internal/requestLogs", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("backend respondeu %d", resp.StatusCode)
	}
	return nil
}
