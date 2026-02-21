package domain

import "time"

// These are correlation IDs — opaque references to aggregates in other contexts.
type AppointmentID string
type InspectorID string
type ClientID string

type AppointmentStatus string

const (
	AppointmentScheduled_  AppointmentStatus = "scheduled"
	AppointmentInProgress  AppointmentStatus = "in_progress"
	AppointmentCompleted   AppointmentStatus = "completed"
	AppointmentCancelled_  AppointmentStatus = "cancelled"
)

// PropertyAddress is the physical location of the property being inspected.
type PropertyAddress struct {
	Street  string
	City    string
	State   string
	Zip     string
	Country string
}

// Appointment is the aggregate root for the scheduling context. It represents
// a booked home inspection at a specific property and time.
type Appointment struct {
	ID                   AppointmentID
	InspectorID          InspectorID
	ClientID             ClientID
	Property             PropertyAddress
	ScheduledAt          time.Time
	EstimatedDurationMin int
	Status               AppointmentStatus
	Notes                string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	events               []any
}

// NewAppointment constructs a valid Appointment in Scheduled status.
func NewAppointment(
	id AppointmentID,
	inspectorID InspectorID,
	clientID ClientID,
	property PropertyAddress,
	scheduledAt time.Time,
	durationMin int,
	notes string,
	now time.Time,
) (*Appointment, error) {
	if !scheduledAt.After(now) {
		return nil, ErrPastScheduledTime
	}
	a := &Appointment{
		ID:                   id,
		InspectorID:          inspectorID,
		ClientID:             clientID,
		Property:             property,
		ScheduledAt:          scheduledAt,
		EstimatedDurationMin: durationMin,
		Status:               AppointmentScheduled_,
		Notes:                notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	a.record(AppointmentScheduled{
		AppointmentID: id,
		InspectorID:   inspectorID,
		ClientID:      clientID,
		OccurredAt:    now,
	})
	return a, nil
}

// Start transitions the appointment to InProgress when the inspector arrives.
func (a *Appointment) Start(now time.Time) error {
	if a.Status != AppointmentScheduled_ {
		return ErrInvalidTransition
	}
	a.Status = AppointmentInProgress
	a.UpdatedAt = now
	return nil
}

// Complete transitions the appointment to Completed.
func (a *Appointment) Complete(now time.Time) error {
	if a.Status != AppointmentInProgress {
		return ErrInvalidTransition
	}
	a.Status = AppointmentCompleted
	a.UpdatedAt = now
	return nil
}

// Cancel transitions the appointment to Cancelled. Completed appointments
// cannot be cancelled.
func (a *Appointment) Cancel(now time.Time) error {
	if a.Status == AppointmentCompleted {
		return ErrCannotCancelCompleted
	}
	a.Status = AppointmentCancelled_
	a.UpdatedAt = now
	a.record(AppointmentCancelled{AppointmentID: a.ID, OccurredAt: now})
	return nil
}

// Reschedule updates the scheduled time. Only allowed while in Scheduled status.
func (a *Appointment) Reschedule(newTime time.Time, now time.Time) error {
	if a.Status != AppointmentScheduled_ {
		return ErrInvalidTransition
	}
	if !newTime.After(now) {
		return ErrPastScheduledTime
	}
	a.ScheduledAt = newTime
	a.UpdatedAt = now
	a.record(AppointmentRescheduled{
		AppointmentID:  a.ID,
		NewScheduledAt: newTime,
		OccurredAt:     now,
	})
	return nil
}

func (a *Appointment) Events() []any { return a.events }
func (a *Appointment) ClearEvents()  { a.events = nil }
func (a *Appointment) record(e any)  { a.events = append(a.events, e) }
