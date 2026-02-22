// Package recorder provides the SyncRecorder that writes mutation audit rows
// into the sync_records table within the same SQL transaction as the mutation.
//
// Usage: inject a *Recorder into each SQLite repository. Calling Record on a
// nil *Recorder is safe — all methods become no-ops, so repos do not need
// extra nil guards.
package recorder

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	syncdomain "github.com/bejayjones/juno/internal/sync/domain"
	"github.com/bejayjones/juno/pkg/id"
)

// Recorder writes sync_records rows inside an existing SQL transaction.
// A nil *Recorder is safe to call — all methods are no-ops.
type Recorder struct {
	clock *syncdomain.LamportClock
}

// New creates a Recorder backed by the given Lamport clock.
func New(clock *syncdomain.LamportClock) *Recorder {
	return &Recorder{clock: clock}
}

// Record ticks the Lamport clock, serialises payload to JSON, and inserts a
// sync_records row inside tx.
//
// operation must be "upsert" or "delete".
// For deletions, pass nil as payload (stored as "{}").
func (r *Recorder) Record(ctx context.Context, tx *sql.Tx, table, recordID, operation string, payload any) error {
	if r == nil {
		return nil
	}

	lc := r.clock.Tick()

	rawPayload := "{}"
	if payload != nil && operation != "delete" {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("sync recorder: marshal payload: %w", err)
		}
		rawPayload = string(b)
	}

	_, err := tx.ExecContext(ctx, `
		INSERT INTO sync_records (id, table_name, record_id, operation, payload, lamport_clock, synced, created_at)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?)
	`, id.New(), table, recordID, operation, rawPayload, lc, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("sync recorder: insert sync_record: %w", err)
	}
	return nil
}
