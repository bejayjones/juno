// Package db provides the database connection wrapper and migration runner
// used by all bounded-context repository implementations.
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Tx is an alias for *sql.Tx, exported so callers don't need to import database/sql.
type Tx = sql.Tx

// DB wraps *sql.DB with Juno-specific helpers (transactions, migrations).
type DB struct {
	*sql.DB
	driver string
}

// Open creates and validates a database connection for the given driver and DSN.
// For SQLite, WAL mode, foreign-key enforcement, and a busy timeout are
// configured automatically.
func Open(driver, dsn string) (*DB, error) {
	sqlDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s db: %w", driver, err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping %s db: %w", driver, err)
	}

	db := &DB{DB: sqlDB, driver: driver}

	if driver == "sqlite" {
		if err := db.configureSQLite(); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// Driver returns the database driver name ("sqlite" or "postgres").
func (db *DB) Driver() string { return db.driver }

func (db *DB) configureSQLite() error {
	// WAL mode allows concurrent reads while a write is in progress.
	// Foreign keys must be enabled per-connection in SQLite.
	// Busy timeout prevents immediate "database is locked" errors under
	// light write contention.
	pragmas := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA synchronous = NORMAL`,
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("sqlite pragma %q: %w", p, err)
		}
	}
	// SQLite supports only one concurrent writer; a pool size of 1 avoids
	// "database is locked" errors without needing external serialization.
	db.SetMaxOpenConns(1)
	return nil
}
