package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migrations/sqlite/*.sql
var sqliteMigrations embed.FS

// Migrate applies all unapplied SQL migrations for the given driver in
// lexicographic filename order. Migration filenames must be prefixed with a
// zero-padded sequence number (e.g. 001_initial_schema.sql) to guarantee
// consistent ordering. Migrations are tracked in a schema_migrations table
// and are idempotent — already-applied migrations are skipped.
func (db *DB) Migrate(ctx context.Context) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT    NOT NULL PRIMARY KEY,
			applied_at INTEGER NOT NULL
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	var (
		fsys embed.FS
		dir  string
	)
	switch db.driver {
	case "sqlite":
		fsys = sqliteMigrations
		dir = "migrations/sqlite"
	default:
		return fmt.Errorf("no embedded migrations for driver %q", db.driver)
	}

	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("read migration directory: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")

		applied, err := db.migrationApplied(ctx, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		content, err := fs.ReadFile(fsys, dir+"/"+entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if err := db.WithTx(ctx, func(tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, string(content)); err != nil {
				return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
			}
			_, err = tx.ExecContext(ctx,
				`INSERT INTO schema_migrations (version, applied_at) VALUES (?, unixepoch())`,
				version,
			)
			return err
		}); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}

		slog.Info("migration applied", "version", version)
	}

	return nil
}

func (db *DB) migrationApplied(ctx context.Context, version string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version,
	).Scan(&count)
	return count > 0, err
}
