// Package sqlite provides the SQLite-backed repository for the sync context.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/bejayjones/juno/internal/platform/db"
	syncdomain "github.com/bejayjones/juno/internal/sync/domain"
)

// SyncRepository persists SyncRecord rows in the sync_records table.
type SyncRepository struct {
	db *db.DB
}

func NewSyncRepository(database *db.DB) *SyncRepository {
	return &SyncRepository{db: database}
}

// Save inserts or updates a sync record. Records coming from a remote push
// have Synced=true; locally generated records have Synced=false.
func (r *SyncRepository) Save(ctx context.Context, record *syncdomain.SyncRecord) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sync_records (id, table_name, record_id, operation, payload, lamport_clock, synced, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			synced        = excluded.synced,
			lamport_clock = excluded.lamport_clock
	`,
		record.ID, record.TableName, record.RecordID,
		record.Operation, record.Payload, record.LamportClock,
		boolToInt(record.Synced), record.CreatedAt.Unix(),
	)
	return err
}

// FindUnsynced returns up to limit records that have not yet been pushed.
func (r *SyncRepository) FindUnsynced(ctx context.Context, limit int) ([]*syncdomain.SyncRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, table_name, record_id, operation, payload, lamport_clock, synced, created_at
		FROM sync_records WHERE synced = 0
		ORDER BY lamport_clock ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecords(rows)
}

// MarkSynced marks the given IDs as pushed to the cloud.
func (r *SyncRepository) MarkSynced(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	_, err := r.db.ExecContext(ctx,
		fmt.Sprintf(`UPDATE sync_records SET synced = 1 WHERE id IN (%s)`, placeholders),
		args...,
	)
	return err
}

// FindSince returns records with lamport_clock strictly greater than since.
func (r *SyncRepository) FindSince(ctx context.Context, since int64) ([]*syncdomain.SyncRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, table_name, record_id, operation, payload, lamport_clock, synced, created_at
		FROM sync_records WHERE lamport_clock > ?
		ORDER BY lamport_clock ASC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRecords(rows)
}

// MaxClock returns the highest lamport_clock value (0 if the table is empty).
func (r *SyncRepository) MaxClock(ctx context.Context) (int64, error) {
	var max sql.NullInt64
	if err := r.db.QueryRowContext(ctx, `SELECT MAX(lamport_clock) FROM sync_records`).Scan(&max); err != nil {
		return 0, err
	}
	if !max.Valid {
		return 0, nil
	}
	return max.Int64, nil
}

// CountUnsynced returns the number of pending (unsynced) records.
func (r *SyncRepository) CountUnsynced(ctx context.Context) (int, error) {
	var n int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sync_records WHERE synced = 0`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

func scanRecords(rows *sql.Rows) ([]*syncdomain.SyncRecord, error) {
	var records []*syncdomain.SyncRecord
	for rows.Next() {
		var (
			rec       syncdomain.SyncRecord
			synced    int
			createdAt int64
		)
		if err := rows.Scan(
			&rec.ID, &rec.TableName, &rec.RecordID,
			&rec.Operation, &rec.Payload, &rec.LamportClock,
			&synced, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan sync_record: %w", err)
		}
		rec.Synced = synced != 0
		rec.CreatedAt = time.Unix(createdAt, 0).UTC()
		records = append(records, &rec)
	}
	return records, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
