package handlers

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// openTestDB abre o Postgres real indicado em HW_TEST_PG (mesmo DSN do .env).
// Sem a variável o teste é pulado, então `go test ./...` não exige banco.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("HW_TEST_PG")
	if dsn == "" {
		t.Skip("HW_TEST_PG não definido: pulando teste com Postgres real")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("conectar no Postgres: %v", err)
	}
	return db
}

func mustExec(t *testing.T, db *gorm.DB, sql string, args ...any) {
	t.Helper()
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

// TestPostgresArrays exercita o que a produção quebrava: o driver devolve
// `uuid[]`/`int[]` como literal `{...}` e o GORM precisa ler com Scanner.
// Cobre o SELECT dos models, o Create do Lobby com turns vazio e o UpsertTurn
// repetido. Usa usuário e partidas de teste e apaga tudo no Cleanup.
func TestPostgresArrays(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	userID := uuid.NewString()
	mustExec(t, db, `INSERT INTO users (id, email, provider, username) VALUES (?::uuid, ?, 'teste', 'teste-arrays')`,
		userID, userID+"@teste.local")
	t.Cleanup(func() {
		db.Exec(`DELETE FROM bug_reports WHERE user_id = ?::uuid`, userID)
		db.Exec(`DELETE FROM match_replays WHERE player_ids @> ARRAY[?]::uuid[]`, userID)
		db.Exec(`DELETE FROM users WHERE id = ?::uuid`, userID)
	})

	t.Run("GORM lê uuid[] e int[]", func(t *testing.T) {
		matchID := uuid.NewString()
		t.Cleanup(func() {
			db.Exec(`DELETE FROM match_replays WHERE match_id = ?::uuid`, matchID)
			db.Exec(`DELETE FROM bug_reports WHERE match_id = ?::uuid`, matchID)
		})

		// Semeia com o literal cru do Postgres, fora do GORM.
		playerA, playerB := uuid.NewString(), uuid.NewString()
		mustExec(t, db, `
INSERT INTO match_replays (match_id, room_id, engine_version, started_at, ended_at, player_ids, data)
VALUES (?::uuid, 'r1', 1, now(), now(), '{`+playerA+`,`+playerB+`}'::uuid[], ?)`,
			matchID, []byte{1, 2, 3})

		var replay models.MatchReplay
		if err := db.WithContext(ctx).First(&replay, "match_id = ?::uuid", matchID).Error; err != nil {
			t.Fatalf("SELECT do replay: %v", err)
		}
		if want := []string{playerA, playerB}; !reflect.DeepEqual([]string(replay.PlayerIDs), want) {
			t.Fatalf("player_ids = %v, esperado %v", replay.PlayerIDs, want)
		}

		mustExec(t, db, `INSERT INTO bug_reports (user_id, source, match_id, turns) VALUES (?::uuid, 'partida', ?::uuid, '{1,2}'::int[])`,
			userID, matchID)

		var report models.BugReport
		if err := db.WithContext(ctx).First(&report, "match_id = ?::uuid", matchID).Error; err != nil {
			t.Fatalf("SELECT do relato: %v", err)
		}
		if want := []int32{1, 2}; !reflect.DeepEqual([]int32(report.Turns), want) {
			t.Fatalf("turns = %v, esperado %v", report.Turns, want)
		}
	})

	t.Run("Create do Lobby grava turns vazio", func(t *testing.T) {
		store := gormBugReportStore{db: db}
		cat, desc := "outro", "descricao suficiente para o relato"
		report := &models.BugReport{
			UserID:      &userID,
			Source:      "lobby",
			Category:    &cat,
			Description: &desc,
			Context:     json.RawMessage(`{"tela":"lobby"}`),
			Status:      "novo",
		}

		if err := store.Create(ctx, report); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if report.ID == 0 || report.CreatedAt.IsZero() || report.UpdatedAt.IsZero() {
			t.Fatalf("RETURNING não preencheu o relato: %+v", report)
		}

		var got models.BugReport
		if err := db.WithContext(ctx).First(&got, "id = ?", report.ID).Error; err != nil {
			t.Fatalf("SELECT do relato criado: %v", err)
		}
		if got.Turns == nil || len(got.Turns) != 0 {
			t.Fatalf("turns = %v, esperado {}", got.Turns)
		}
		if string(got.Context) == "" {
			t.Fatal("context não gravado")
		}
	})

	t.Run("UpsertTurn repetido devolve a mesma linha", func(t *testing.T) {
		store := gormBugReportStore{db: db}
		matchID := uuid.NewString()

		id1, turns1, created1, err := store.UpsertTurn(ctx, userID, "teste-arrays", matchID, "127.0.0.1", "teste", 3)
		if err != nil {
			t.Fatalf("primeiro UpsertTurn: %v", err)
		}
		if !created1 {
			t.Fatal("primeiro UpsertTurn deveria criar o relato")
		}
		if want := []int32{3}; !reflect.DeepEqual(turns1, want) {
			t.Fatalf("turns = %v, esperado %v", turns1, want)
		}

		id2, turns2, created2, err := store.UpsertTurn(ctx, userID, "teste-arrays", matchID, "127.0.0.1", "teste", 3)
		if err != nil {
			t.Fatalf("segundo UpsertTurn: %v", err)
		}
		if created2 {
			t.Fatal("segundo UpsertTurn não deveria criar outra linha")
		}
		if id1 != id2 {
			t.Fatalf("id mudou: %d → %d", id1, id2)
		}
		if want := []int32{3}; !reflect.DeepEqual(turns2, want) {
			t.Fatalf("turns duplicou: %v, esperado %v", turns2, want)
		}

		var n int64
		if err := db.WithContext(ctx).Model(&models.BugReport{}).
			Where("user_id = ?::uuid AND match_id = ?::uuid", userID, matchID).
			Count(&n).Error; err != nil {
			t.Fatalf("count: %v", err)
		}
		if n != 1 {
			t.Fatalf("esperado 1 relato para o turno, veio %d", n)
		}
	})
}
