package domain

import "time"

type ReportGenerated struct {
	ReportID     ReportID
	InspectionID InspectionID
	OccurredAt   time.Time
}

type ReportWasFinalized struct {
	ReportID   ReportID
	OccurredAt time.Time
}

type DeliveryQueued struct {
	ReportID       ReportID
	DeliveryID     DeliveryID
	RecipientEmail string
	OccurredAt     time.Time
}

type DeliverySucceeded struct {
	ReportID   ReportID
	DeliveryID DeliveryID
	OccurredAt time.Time
}

type DeliveryFailed struct {
	ReportID      ReportID
	DeliveryID    DeliveryID
	FailureReason string
	OccurredAt    time.Time
}
