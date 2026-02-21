package domain

import "context"

// ReportRepository is the persistence contract for the reporting context.
type ReportRepository interface {
	Save(ctx context.Context, report *Report) error
	FindByID(ctx context.Context, id ReportID) (*Report, error)
	FindByInspection(ctx context.Context, inspectionID InspectionID) (*Report, error)
	FindByInspector(ctx context.Context, inspectorID InspectorID, filter ReportFilter) ([]*Report, error)
	Delete(ctx context.Context, id ReportID) error
}

type ReportFilter struct {
	Status *ReportStatus
	Limit  int
	Offset int
}
