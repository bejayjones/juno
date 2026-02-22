package application_test

import (
	"context"
	"testing"
	"time"

	syncapp "github.com/bejayjones/juno/internal/sync/application"
	syncdomain "github.com/bejayjones/juno/internal/sync/domain"
)

// ── fake repository ──────────────────────────────────────────────────────────

type fakeSyncRepo struct {
	records []*syncdomain.SyncRecord
}

func (r *fakeSyncRepo) Save(_ context.Context, rec *syncdomain.SyncRecord) error {
	for i, existing := range r.records {
		if existing.ID == rec.ID {
			r.records[i] = rec
			return nil
		}
	}
	r.records = append(r.records, rec)
	return nil
}

func (r *fakeSyncRepo) FindUnsynced(_ context.Context, limit int) ([]*syncdomain.SyncRecord, error) {
	var out []*syncdomain.SyncRecord
	for _, rec := range r.records {
		if !rec.Synced {
			out = append(out, rec)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *fakeSyncRepo) MarkSynced(_ context.Context, ids []string) error {
	set := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	for _, rec := range r.records {
		if _, ok := set[rec.ID]; ok {
			rec.Synced = true
		}
	}
	return nil
}

func (r *fakeSyncRepo) FindSince(_ context.Context, since int64) ([]*syncdomain.SyncRecord, error) {
	var out []*syncdomain.SyncRecord
	for _, rec := range r.records {
		if rec.LamportClock > since {
			out = append(out, rec)
		}
	}
	return out, nil
}

func (r *fakeSyncRepo) MaxClock(_ context.Context) (int64, error) {
	var max int64
	for _, rec := range r.records {
		if rec.LamportClock > max {
			max = rec.LamportClock
		}
	}
	return max, nil
}

func (r *fakeSyncRepo) CountUnsynced(_ context.Context) (int, error) {
	n := 0
	for _, rec := range r.records {
		if !rec.Synced {
			n++
		}
	}
	return n, nil
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestGetStatus_Empty(t *testing.T) {
	repo := &fakeSyncRepo{}
	clock := syncdomain.NewLamportClock(0)
	svc := syncapp.NewSyncService(repo, clock)

	view, err := svc.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if view.PendingCount != 0 {
		t.Errorf("pending count = %d, want 0", view.PendingCount)
	}
	if view.CurrentClock != 0 {
		t.Errorf("current clock = %d, want 0", view.CurrentClock)
	}
}

func TestHandlePush_AdvancesClock(t *testing.T) {
	repo := &fakeSyncRepo{}
	clock := syncdomain.NewLamportClock(0)
	svc := syncapp.NewSyncService(repo, clock)

	records := []*syncdomain.SyncRecord{
		{ID: "r1", TableName: "inspectors", RecordID: "i1", Operation: "upsert", Payload: "{}", LamportClock: 5, CreatedAt: time.Now()},
		{ID: "r2", TableName: "inspectors", RecordID: "i2", Operation: "upsert", Payload: "{}", LamportClock: 10, CreatedAt: time.Now()},
	}

	result, err := svc.HandlePush(context.Background(), records)
	if err != nil {
		t.Fatalf("HandlePush: %v", err)
	}
	if result.Applied != 2 {
		t.Errorf("applied = %d, want 2", result.Applied)
	}
	// Clock should be > 10 after witnessing the remote value.
	if clock.Current() <= 10 {
		t.Errorf("clock = %d, want > 10", clock.Current())
	}
	// Pushed records should be saved as synced.
	for _, rec := range repo.records {
		if !rec.Synced {
			t.Errorf("record %s should be marked synced", rec.ID)
		}
	}
}

func TestHandlePull_FiltersByClock(t *testing.T) {
	repo := &fakeSyncRepo{
		records: []*syncdomain.SyncRecord{
			{ID: "r1", TableName: "appointments", RecordID: "a1", Operation: "upsert", Payload: "{}", LamportClock: 3, Synced: true, CreatedAt: time.Now()},
			{ID: "r2", TableName: "appointments", RecordID: "a2", Operation: "upsert", Payload: "{}", LamportClock: 7, Synced: true, CreatedAt: time.Now()},
			{ID: "r3", TableName: "appointments", RecordID: "a3", Operation: "upsert", Payload: "{}", LamportClock: 12, Synced: true, CreatedAt: time.Now()},
		},
	}
	clock := syncdomain.NewLamportClock(12)
	svc := syncapp.NewSyncService(repo, clock)

	result, err := svc.HandlePull(context.Background(), 5)
	if err != nil {
		t.Fatalf("HandlePull: %v", err)
	}
	if len(result.Records) != 2 {
		t.Errorf("records = %d, want 2 (clocks 7 and 12)", len(result.Records))
	}
}

func TestGetStatus_CountsPending(t *testing.T) {
	repo := &fakeSyncRepo{
		records: []*syncdomain.SyncRecord{
			{ID: "r1", LamportClock: 1, Synced: false},
			{ID: "r2", LamportClock: 2, Synced: true},
			{ID: "r3", LamportClock: 3, Synced: false},
		},
	}
	clock := syncdomain.NewLamportClock(3)
	svc := syncapp.NewSyncService(repo, clock)

	view, err := svc.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus: %v", err)
	}
	if view.PendingCount != 2 {
		t.Errorf("pending = %d, want 2", view.PendingCount)
	}
}
