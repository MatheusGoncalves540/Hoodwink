package services

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"gorm.io/gorm"
)

const (
	// requestLogChannelSize é o buffer do canal de gravação assíncrona (plano 04).
	requestLogChannelSize = 1000
	// requestLogBatchSize é o tamanho do lote antes de gravar (plano 04).
	requestLogBatchSize = 100
	// requestLogFlushInterval é o intervalo máximo entre gravações de lote.
	requestLogFlushInterval = 2 * time.Second
	// requestLogDropLogInterval é o intervalo para logar descartes acumulados.
	requestLogDropLogInterval = time.Minute
	// requestLogWriteTimeout limita cada INSERT (drenagem do SIGTERM usa 3s).
	requestLogWriteTimeout = 3 * time.Second
)

// RequestLogService grava linhas do coletor de forma assíncrona: canal com
// buffer, lote a cada 2s ou 100 linhas. Canal cheio descarta e, a cada minuto
// com descarte, loga `requestlog dropped=N`. Close() drena e grava o último lote.
type RequestLogService struct {
	db        *gorm.DB
	ch        chan *models.RequestLog
	done      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once
	dropped   atomic.Int64
	// insertFn é substituível em testes para observar os lotes sem Postgres.
	insertFn func([]*models.RequestLog)
}

func NewRequestLogService(db *gorm.DB) *RequestLogService {
	s := newRequestLogService(db)
	s.insertFn = s.insert
	s.start()
	return s
}

// NewRequestLogServiceForTest cria um serviço cuja gravação é desviada para fn,
// permitindo observar os lotes sem Postgres. fn nunca é nil.
func NewRequestLogServiceForTest(fn func([]*models.RequestLog)) *RequestLogService {
	s := newRequestLogService(nil)
	s.insertFn = fn
	s.start()
	return s
}

func newRequestLogService(db *gorm.DB) *RequestLogService {
	return &RequestLogService{
		db:   db,
		ch:   make(chan *models.RequestLog, requestLogChannelSize),
		done: make(chan struct{}),
	}
}

func (s *RequestLogService) start() {
	s.wg.Add(1)
	go s.loop()
}

// Enqueue enfileira uma linha para gravação. Se o canal estiver cheio, descarta
// e registra o descarte (logado uma vez por minuto pelo loop).
func (s *RequestLogService) Enqueue(entry *models.RequestLog) {
	if entry == nil {
		return
	}
	select {
	case s.ch <- entry:
	default:
		s.dropped.Add(1)
	}
}

// InsertBatch grava de forma síncrona (usado pelo ingest do gameServer).
func (s *RequestLogService) InsertBatch(ctx context.Context, entries []*models.RequestLog) error {
	if len(entries) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Create(&entries).Error
}

// Close sinaliza o loop para drenar o canal e gravar o último lote, aguardando
// o término. Cada INSERT usa requestLogWriteTimeout (3s) como teto. É idempotente.
func (s *RequestLogService) Close() {
	s.closeOnce.Do(func() {
		close(s.done)
		s.wg.Wait()
	})
}

func (s *RequestLogService) loop() {
	defer s.wg.Done()

	batch := make([]*models.RequestLog, 0, requestLogBatchSize)
	ticker := time.NewTicker(requestLogFlushInterval)
	dropTicker := time.NewTicker(requestLogDropLogInterval)
	defer ticker.Stop()
	defer dropTicker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		s.insertFn(batch)
		batch = batch[:0]
	}

	for {
		select {
		case entry := <-s.ch:
			batch = append(batch, entry)
			if len(batch) >= requestLogBatchSize {
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
			// Drena o que ainda estiver no canal antes de encerrar.
			for {
				select {
				case entry := <-s.ch:
					batch = append(batch, entry)
					if len(batch) >= requestLogBatchSize {
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

func (s *RequestLogService) insert(entries []*models.RequestLog) {
	ctx, cancel := context.WithTimeout(context.Background(), requestLogWriteTimeout)
	defer cancel()
	if err := s.db.WithContext(ctx).Create(&entries).Error; err != nil {
		log.Printf("requestlog: falha ao gravar lote de %d linhas: %v", len(entries), err)
	}
}
