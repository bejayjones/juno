package domain

import "errors"

var (
	ErrInspectorNotFound = errors.New("inspector not found")
	ErrCompanyNotFound   = errors.New("company not found")
	ErrClientNotFound    = errors.New("client not found")
	ErrEmailTaken        = errors.New("email address is already in use")
	ErrInvalidEmail      = errors.New("invalid email address")
	ErrInvalidRole       = errors.New("invalid inspector role")
)
