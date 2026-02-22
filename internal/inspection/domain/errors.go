package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInspectionNotFound  = errors.New("inspection not found")
	ErrInspectionCompleted = errors.New("inspection is already completed")
	ErrItemNotFound        = errors.New("inspection item not found")
	ErrFindingNotFound     = errors.New("finding not found")
	ErrPhotoNotFound       = errors.New("photo not found")
	ErrInvalidSystemType   = errors.New("invalid system type")
	ErrNIReasonRequired    = errors.New("reason is required when status is NotInspected")
	ErrInvalidMimeType     = errors.New("unsupported image type; accepted: JPEG, PNG, HEIC")
	ErrPhotoTooLarge       = errors.New("photo exceeds 20 MB limit")
)

// ValidationError is returned by Complete when required fields are unfilled.
type ValidationError struct {
	Fields []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("inspection incomplete: %s", strings.Join(e.Fields, "; "))
}
