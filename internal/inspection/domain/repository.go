package domain

import (
	"context"
	"time"
)

// InspectionRepository is the persistence contract for the inspection bounded context.
// Implementations live in infrastructure/sqlite and infrastructure/postgres.
type InspectionRepository interface {
	Save(ctx context.Context, inspection *Inspection) error
	FindByID(ctx context.Context, id InspectionID) (*Inspection, error)
	FindByAppointmentID(ctx context.Context, appointmentID AppointmentID) (*Inspection, error)
	FindByInspector(ctx context.Context, inspectorID InspectorID, filter InspectionFilter) ([]*Inspection, error)
	// FindPhotoMeta returns the storage path and MIME type for a photo by ID.
	// This is used to serve photo data without loading the full inspection aggregate.
	FindPhotoMeta(ctx context.Context, photoID PhotoID) (storagePath, mimeType string, err error)
	Delete(ctx context.Context, id InspectionID) error
}

type InspectionFilter struct {
	Status   *InspectionStatus
	FromDate *time.Time
	ToDate   *time.Time
	Limit    int
	Offset   int
}
