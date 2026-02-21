package domain

import "errors"

var (
	ErrAppointmentNotFound    = errors.New("appointment not found")
	ErrInvalidTransition      = errors.New("status transition is not allowed")
	ErrCannotCancelCompleted  = errors.New("cannot cancel a completed appointment")
	ErrPastScheduledTime      = errors.New("scheduled time cannot be in the past")
)
