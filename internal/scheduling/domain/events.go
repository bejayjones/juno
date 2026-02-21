package domain

import "time"

type AppointmentScheduled struct {
	AppointmentID AppointmentID
	InspectorID   InspectorID
	ClientID      ClientID
	OccurredAt    time.Time
}

type AppointmentCancelled struct {
	AppointmentID AppointmentID
	OccurredAt    time.Time
}

type AppointmentRescheduled struct {
	AppointmentID  AppointmentID
	NewScheduledAt time.Time
	OccurredAt     time.Time
}
