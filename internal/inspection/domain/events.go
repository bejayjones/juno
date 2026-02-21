package domain

import "time"

type InspectionStarted struct {
	InspectionID  InspectionID
	AppointmentID AppointmentID
	InspectorID   InspectorID
	OccurredAt    time.Time
}

type ItemStatusUpdated struct {
	InspectionID InspectionID
	SystemType   SystemType
	ItemKey      ItemKey
	OldStatus    ItemStatus
	NewStatus    ItemStatus
	OccurredAt   time.Time
}

type DeficiencyRecorded struct {
	InspectionID InspectionID
	SystemType   SystemType
	ItemKey      ItemKey
	FindingID    FindingID
	OccurredAt   time.Time
}

type InspectionCompleted struct {
	InspectionID InspectionID
	OccurredAt   time.Time
}
