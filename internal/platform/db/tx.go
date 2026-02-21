package db

import (
	"context"
	"database/sql"
	"fmt"
)

// WithTx runs fn inside a database transaction. The transaction is committed
// if fn returns nil, or rolled back if fn returns an error or panics.
func (db *DB) WithTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %w; rollback also failed: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
