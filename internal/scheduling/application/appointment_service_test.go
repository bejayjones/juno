package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bejayjones/juno/internal/scheduling/application"
	"github.com/bejayjones/juno/internal/scheduling/domain"
	"github.com/bejayjones/juno/pkg/clock"
)

// fakeAppointmentRepo is an in-memory AppointmentRepository for tests.
type fakeAppointmentRepo struct {
	byID map[domain.AppointmentID]*domain.Appointment
}

func newFakeAppointmentRepo() *fakeAppointmentRepo {
	return &fakeAppointmentRepo{byID: make(map[domain.AppointmentID]*domain.Appointment)}
}

func (r *fakeAppointmentRepo) Save(_ context.Context, a *domain.Appointment) error {
	r.byID[a.ID] = a
	return nil
}

func (r *fakeAppointmentRepo) FindByID(_ context.Context, id domain.AppointmentID) (*domain.Appointment, error) {
	if a, ok := r.byID[id]; ok {
		return a, nil
	}
	return nil, domain.ErrAppointmentNotFound
}

func (r *fakeAppointmentRepo) FindByInspector(_ context.Context, inspectorID domain.InspectorID, filter domain.AppointmentFilter) ([]*domain.Appointment, error) {
	var result []*domain.Appointment
	for _, a := range r.byID {
		if a.InspectorID != inspectorID {
			continue
		}
		if filter.Status != nil && a.Status != *filter.Status {
			continue
		}
		if filter.FromDate != nil && a.ScheduledAt.Before(*filter.FromDate) {
			continue
		}
		if filter.ToDate != nil && a.ScheduledAt.After(*filter.ToDate) {
			continue
		}
		result = append(result, a)
	}
	return result, nil
}

func (r *fakeAppointmentRepo) Delete(_ context.Context, id domain.AppointmentID) error {
	delete(r.byID, id)
	return nil
}

var testNow = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
var futureTime = testNow.Add(48 * time.Hour)

func newTestService() *application.AppointmentService {
	return application.NewAppointmentService(
		newFakeAppointmentRepo(),
		clock.Fixed(testNow),
	)
}

func TestSchedule_Success(t *testing.T) {
	svc := newTestService()

	view, err := svc.Schedule(context.Background(), application.ScheduleInput{
		InspectorID: "insp-1",
		ClientID:    "client-1",
		Street:      "123 Main St",
		City:        "Austin",
		State:       "TX",
		Zip:         "78701",
		ScheduledAt: futureTime,
		DurationMin: 180,
	})
	if err != nil {
		t.Fatalf("Schedule: %v", err)
	}

	if view.InspectorID != "insp-1" {
		t.Errorf("inspector_id = %q", view.InspectorID)
	}
	if view.Status != "scheduled" {
		t.Errorf("status = %q, want scheduled", view.Status)
	}
	if view.DurationMin != 180 {
		t.Errorf("duration = %d, want 180", view.DurationMin)
	}
	if view.ID == "" {
		t.Error("expected non-empty id")
	}
}

func TestSchedule_PastTimeFails(t *testing.T) {
	svc := newTestService()
	_, err := svc.Schedule(context.Background(), application.ScheduleInput{
		InspectorID: "insp-1",
		ClientID:    "client-1",
		Street:      "123 Main St",
		City:        "Austin",
		ScheduledAt: testNow.Add(-1 * time.Hour), // in the past
	})
	if !errors.Is(err, domain.ErrPastScheduledTime) {
		t.Errorf("want ErrPastScheduledTime, got %v", err)
	}
}

func TestSchedule_DefaultDuration(t *testing.T) {
	svc := newTestService()
	view, err := svc.Schedule(context.Background(), application.ScheduleInput{
		InspectorID: "insp-1",
		ClientID:    "client-1",
		Street:      "1 Oak Ave",
		City:        "Dallas",
		ScheduledAt: futureTime,
		// DurationMin omitted — should default to 120
	})
	if err != nil {
		t.Fatalf("Schedule: %v", err)
	}
	if view.DurationMin != 120 {
		t.Errorf("duration = %d, want 120", view.DurationMin)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetByID(context.Background(), domain.AppointmentID("nonexistent"))
	if !errors.Is(err, domain.ErrAppointmentNotFound) {
		t.Errorf("want ErrAppointmentNotFound, got %v", err)
	}
}

func TestList_FilterByStatus(t *testing.T) {
	svc := newTestService()

	// Create two scheduled appointments.
	for range 2 {
		_, _ = svc.Schedule(context.Background(), application.ScheduleInput{
			InspectorID: "insp-1",
			ClientID:    "client-1",
			Street:      "1 Main",
			City:        "Austin",
			ScheduledAt: futureTime,
		})
	}

	st := domain.AppointmentScheduled_
	views, err := svc.List(context.Background(), application.ListInput{
		InspectorID: "insp-1",
		Status:      &st,
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(views) != 2 {
		t.Errorf("want 2 appointments, got %d", len(views))
	}
}

func TestUpdate_Reschedule(t *testing.T) {
	svc := newTestService()
	created, _ := svc.Schedule(context.Background(), application.ScheduleInput{
		InspectorID: "insp-1",
		ClientID:    "client-1",
		Street:      "1 Main",
		City:        "Austin",
		ScheduledAt: futureTime,
	})

	newTime := futureTime.Add(24 * time.Hour)
	updated, err := svc.Update(context.Background(), domain.AppointmentID(created.ID), application.UpdateInput{
		Street:      "2 Oak Ave",
		City:        "Austin",
		DurationMin: 90,
		ScheduledAt: &newTime,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.ScheduledAt != newTime.Unix() {
		t.Errorf("scheduled_at not updated")
	}
	if updated.Property.Street != "2 Oak Ave" {
		t.Errorf("street = %q", updated.Property.Street)
	}
}

func TestCancel_Success(t *testing.T) {
	svc := newTestService()
	created, _ := svc.Schedule(context.Background(), application.ScheduleInput{
		InspectorID: "insp-1",
		ClientID:    "client-1",
		Street:      "1 Main",
		City:        "Austin",
		ScheduledAt: futureTime,
	})

	if err := svc.Cancel(context.Background(), domain.AppointmentID(created.ID)); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	view, _ := svc.GetByID(context.Background(), domain.AppointmentID(created.ID))
	if view.Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", view.Status)
	}
}
