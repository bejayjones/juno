package db_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/bejayjones/juno/internal/platform/db"
)

func TestOpen_SQLite(t *testing.T) {
	database, err := db.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestMigrate_CreatesAllTables(t *testing.T) {
	database := OpenTestDB(t)
	ctx := context.Background()

	want := []string{
		"companies", "inspectors", "inspector_licenses", "clients",
		"appointments",
		"inspections", "system_sections", "system_descriptions",
		"inspection_items", "findings", "photos",
		"reports", "deliveries",
		"sync_records",
		"schema_migrations",
	}

	for _, table := range want {
		var count int
		err := database.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, table,
		).Scan(&count)
		if err != nil {
			t.Errorf("query for table %q: %v", table, err)
			continue
		}
		if count == 0 {
			t.Errorf("table %q was not created", table)
		}
	}
}

func TestMigrate_IsIdempotent(t *testing.T) {
	database, err := db.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := database.Migrate(ctx); err != nil {
			t.Fatalf("Migrate (run %d): %v", i+1, err)
		}
	}
}

func TestWithTx_CommitsOnSuccess(t *testing.T) {
	database := OpenTestDB(t)
	ctx := context.Background()

	err := database.WithTx(ctx, func(tx *db.Tx) error {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO companies (id, name, created_at, updated_at) VALUES ('c1', 'Test Co', 1, 1)`,
		)
		return err
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	var count int
	database.QueryRowContext(ctx, `SELECT COUNT(*) FROM companies WHERE id='c1'`).Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	database := OpenTestDB(t)
	ctx := context.Background()

	_ = database.WithTx(ctx, func(tx *db.Tx) error {
		tx.ExecContext(ctx,
			`INSERT INTO companies (id, name, created_at, updated_at) VALUES ('c2', 'Rollback Co', 1, 1)`,
		)
		return fmt.Errorf("intentional error")
	})

	var count int
	database.QueryRowContext(ctx, `SELECT COUNT(*) FROM companies WHERE id='c2'`).Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 rows after rollback, got %d", count)
	}
}

// OpenTestDB opens an in-memory SQLite database and runs all migrations.
// It registers t.Cleanup to close the connection when the test ends.
// Other packages that need a seeded test DB can copy this helper locally.
func OpenTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("OpenTestDB: %v", err)
	}
	if err := database.Migrate(context.Background()); err != nil {
		t.Fatalf("OpenTestDB migrate: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
