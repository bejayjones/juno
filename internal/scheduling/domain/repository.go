package domain

import (
	"context"
	"time"
)

// AppointmentRepository is the persistence contract for the scheduling context.
type AppointmentRepository interface {
	Save(ctx context.Context, appointment *Appointment) error
	FindByID(ctx context.Context, id AppointmentID) (*Appointment, error)
	FindByInspector(ctx context.Context, inspectorID InspectorID, filter AppointmentFilter) ([]*Appointment, error)
	Delete(ctx context.Context, id AppointmentID) error
}

type AppointmentFilter struct {
	Status   *AppointmentStatus
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
}
