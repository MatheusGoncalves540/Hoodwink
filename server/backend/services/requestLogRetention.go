package services

import (
	"context"
	"log"

	"gorm.io/gorm"
)

// retentionLockKey é a chave do advisory lock xact (constante do plano 04).
const retentionLockKey = 54526370001

// moveLogsSQL move um lote de linhas antigas da tabela ativa para o arquivo
// preservando o id original. Parâmetros são constantes fixas (nunca input).
const moveLogsSQL = `
WITH d AS (
  DELETE FROM request_logs
  WHERE id IN (SELECT id FROM request_logs WHERE at < now() - interval '30 days' LIMIT 10000)
  RETURNING *
)
INSERT INTO request_logs_archive SELECT * FROM d`

const deleteArchiveSQL = `DELETE FROM request_logs_archive WHERE at < now() - interval '6 months'`

// deleteReplaysSQL apaga um lote de replays com mais de 30 dias (plano 05).
const deleteReplaysSQL = `
WITH d AS (
  SELECT match_id FROM match_replays
  WHERE ended_at < now() - interval '30 days'
  LIMIT 10000
)
DELETE FROM match_replays WHERE match_id IN (SELECT match_id FROM d)`

// RunRetention executa a retenção uma vez: arquiva linhas com mais de 30 dias
// (em lotes de 10.000 dentro de uma transação) e apaga do arquivo o que tem
// mais de 6 meses. Usa advisory lock de transação; se outra instância já estiver
// rodando, sai sem erro.
func RunRetention(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked bool
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", retentionLockKey).Scan(&locked).Error; err != nil {
			return err
		}
		if !locked {
			log.Println("requestlog retention: lock já tomado por outra instância, saindo")
			return nil
		}

		if err := runRetentionLoop(func() (int64, error) {
			res := tx.Exec(moveLogsSQL)
			return res.RowsAffected, res.Error
		}); err != nil {
			return err
		}

		// Replays com mais de 30 dias (mesmo advisory lock, em lotes).
		if err := runRetentionLoop(func() (int64, error) {
			res := tx.Exec(deleteReplaysSQL)
			return res.RowsAffected, res.Error
		}); err != nil {
			return err
		}

		return tx.Exec(deleteArchiveSQL).Error
	})
}

// runRetentionLoop chama move até ele relatar 0 linhas movidas (lote a lote).
func runRetentionLoop(move func() (int64, error)) error {
	for {
		n, err := move()
		if err != nil {
			return err
		}
		if n == 0 {
			return nil
		}
	}
}
