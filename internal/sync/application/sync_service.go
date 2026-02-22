package application

import (
	"context"
	"fmt"

	syncdomain "github.com/bejayjones/juno/internal/sync/domain"
)

// SyncService handles push/pull/status operations for the sync protocol.
type SyncService struct {
	repo  syncdomain.Repository
	clock *syncdomain.LamportClock
}

// NewSyncService creates a SyncService. clock must be initialised with the
// maximum lamport_clock value from the local sync_records table so it survives
// restarts.
func NewSyncService(repo syncdomain.Repository, clock *syncdomain.LamportClock) *SyncService {
	return &SyncService{repo: repo, clock: clock}
}

// GetStatus returns the current sync state.
func (s *SyncService) GetStatus(ctx context.Context) (*StatusView, error) {
	n, err := s.repo.CountUnsynced(ctx)
	if err != nil {
		return nil, fmt.Errorf("sync status: count unsynced: %w", err)
	}
	return &StatusView{
		PendingCount: n,
		CurrentClock: s.clock.Current(),
	}, nil
}

// HandlePush stores records received from a remote device. The Lamport clock
// is advanced to stay ahead of any incoming clock values.
func (s *SyncService) HandlePush(ctx context.Context, records []*syncdomain.SyncRecord) (*PushResult, error) {
	for _, rec := range records {
		s.clock.Witness(rec.LamportClock)
		// Mark as synced — the record originated on the remote and is already there.
		rec.Synced = true
		if err := s.repo.Save(ctx, rec); err != nil {
			return nil, fmt.Errorf("sync push: save record %s: %w", rec.ID, err)
		}
	}
	return &PushResult{Applied: len(records)}, nil
}

// HandlePull returns all sync records with a Lamport clock greater than since.
// Remote devices call this to catch up on changes from other devices.
func (s *SyncService) HandlePull(ctx context.Context, since int64) (*PullResult, error) {
	records, err := s.repo.FindSince(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("sync pull: find since %d: %w", since, err)
	}
	views := make([]*SyncRecordView, len(records))
	for i, r := range records {
		views[i] = &SyncRecordView{
			ID:           r.ID,
			TableName:    r.TableName,
			RecordID:     r.RecordID,
			Operation:    r.Operation,
			Payload:      r.Payload,
			LamportClock: r.LamportClock,
			CreatedAt:    r.CreatedAt.Unix(),
		}
	}
	return &PullResult{Records: views}, nil
}
