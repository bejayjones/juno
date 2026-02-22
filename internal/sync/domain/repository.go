// Package domain defines the core types for the sync bounded context.
package domain

import "context"

// Repository is the persistence contract for the sync bounded context.
type Repository interface {
	// Save persists a sync record. Records created locally by the Recorder
	// arrive with Synced=false; records received from a remote push arrive
	// with Synced=true.
	Save(ctx context.Context, record *SyncRecord) error

	// FindUnsynced returns up to limit records that have not yet been pushed
	// to the cloud, ordered by lamport_clock ascending.
	FindUnsynced(ctx context.Context, limit int) ([]*SyncRecord, error)

	// MarkSynced marks the given record IDs as pushed to the cloud.
	MarkSynced(ctx context.Context, ids []string) error

	// FindSince returns all records whose lamport_clock is strictly greater
	// than since, ordered by lamport_clock ascending. Used for the pull protocol.
	FindSince(ctx context.Context, since int64) ([]*SyncRecord, error)

	// MaxClock returns the highest lamport_clock stored locally (0 if empty).
	// Called at startup to seed the LamportClock.
	MaxClock(ctx context.Context) (int64, error)

	// CountUnsynced returns the total number of pending (unsynced) records.
	CountUnsynced(ctx context.Context) (int, error)
}
