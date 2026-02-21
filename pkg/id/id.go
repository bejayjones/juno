package id

import "github.com/google/uuid"

// New generates a new unique identifier.
func New() string {
	return uuid.New().String()
}

// Validate reports whether s is a valid UUID.
func Validate(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}
