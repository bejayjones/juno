package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bejayjones/juno/internal/scheduling/domain"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/id"
)

// AppointmentService handles scheduling operations.
type AppointmentService struct {
	appointments domain.AppointmentRepository
	clock        clock.Clock
}

func NewAppointmentService(appointments domain.AppointmentRepository, clk clock.Clock) *AppointmentService {
	return &AppointmentService{appointments: appointments, clock: clk}
}

// ScheduleInput contains the fields needed to book a new appointment.
type ScheduleInput struct {
	InspectorID string
	ClientID    string
	Street      string
	City        string
	State       string
	Zip         string
	Country     string
	ScheduledAt time.Time
	DurationMin int
	Notes       string
}

// Schedule books a new appointment, returning its view on success.
func (s *AppointmentService) Schedule(ctx context.Context, in ScheduleInput) (AppointmentView, error) {
	if in.DurationMin <= 0 {
		in.DurationMin = 120
	}
	if in.Country == "" {
		in.Country = "US"
	}

	appt, err := domain.NewAppointment(
		domain.AppointmentID(id.New()),
		domain.InspectorID(in.InspectorID),
		domain.ClientID(in.ClientID),
		domain.PropertyAddress{
			Street: in.Street, City: in.City,
			State: in.State, Zip: in.Zip, Country: in.Country,
		},
		in.ScheduledAt,
		in.DurationMin,
		in.Notes,
		s.clock.Now(),
	)
	if err != nil {
		return AppointmentView{}, err
	}

	if err := s.appointments.Save(ctx, appt); err != nil {
		return AppointmentView{}, fmt.Errorf("save appointment: %w", err)
	}
	return toAppointmentView(appt), nil
}

// GetByID returns a single appointment.
func (s *AppointmentService) GetByID(ctx context.Context, id domain.AppointmentID) (AppointmentView, error) {
	appt, err := s.appointments.FindByID(ctx, id)
	if err != nil {
		return AppointmentView{}, err
	}
	return toAppointmentView(appt), nil
}

// ListInput contains filter parameters for listing appointments.
type ListInput struct {
	InspectorID string
	Status      *domain.AppointmentStatus
	FromDate    *time.Time
	ToDate      *time.Time
	Limit       int
	Offset      int
}

// List returns appointments for an inspector, applying optional filters.
func (s *AppointmentService) List(ctx context.Context, in ListInput) ([]AppointmentView, error) {
	appts, err := s.appointments.FindByInspector(ctx, domain.InspectorID(in.InspectorID), domain.AppointmentFilter{
		Status:   in.Status,
		FromDate: in.FromDate,
		ToDate:   in.ToDate,
		Limit:    in.Limit,
		Offset:   in.Offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list appointments: %w", err)
	}
	views := make([]AppointmentView, len(appts))
	for i, a := range appts {
		views[i] = toAppointmentView(a)
	}
	return views, nil
}

// UpdateInput contains the mutable fields for a scheduled appointment.
type UpdateInput struct {
	Street      string
	City        string
	State       string
	Zip         string
	Country     string
	ScheduledAt *time.Time // nil = no change
	DurationMin int
	Notes       string
}

// Update modifies the details of a scheduled appointment.
// Rescheduling (changing ScheduledAt) is only allowed for scheduled appointments.
// Other field changes are allowed for scheduled appointments as well.
func (s *AppointmentService) Update(ctx context.Context, apptID domain.AppointmentID, in UpdateInput) (AppointmentView, error) {
	appt, err := s.appointments.FindByID(ctx, apptID)
	if err != nil {
		return AppointmentView{}, err
	}

	now := s.clock.Now()

	if in.ScheduledAt != nil && !in.ScheduledAt.Equal(appt.ScheduledAt) {
		if err := appt.Reschedule(*in.ScheduledAt, now); err != nil {
			return AppointmentView{}, err
		}
	}

	durationMin := in.DurationMin
	if durationMin <= 0 {
		durationMin = appt.EstimatedDurationMin
	}
	country := in.Country
	if country == "" {
		country = appt.Property.Country
	}

	if err := appt.UpdateDetails(
		domain.PropertyAddress{Street: in.Street, City: in.City, State: in.State, Zip: in.Zip, Country: country},
		durationMin,
		in.Notes,
		now,
	); err != nil {
		return AppointmentView{}, err
	}

	if err := s.appointments.Save(ctx, appt); err != nil {
		return AppointmentView{}, fmt.Errorf("save appointment: %w", err)
	}
	return toAppointmentView(appt), nil
}

// Cancel transitions an appointment to Cancelled status.
func (s *AppointmentService) Cancel(ctx context.Context, apptID domain.AppointmentID) error {
	appt, err := s.appointments.FindByID(ctx, apptID)
	if err != nil {
		return err
	}
	if err := appt.Cancel(s.clock.Now()); err != nil {
		return err
	}
	if err := s.appointments.Save(ctx, appt); err != nil {
		return fmt.Errorf("save appointment: %w", err)
	}
	return nil
}

// IsCancelled reports whether the error is the cannot-cancel-completed sentinel.
func IsCancelled(err error) bool {
	return errors.Is(err, domain.ErrCannotCancelCompleted) || errors.Is(err, domain.ErrInvalidTransition)
}
