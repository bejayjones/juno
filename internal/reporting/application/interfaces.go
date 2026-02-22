package application

import (
	"context"
	"errors"

	inspectiondomain "github.com/bejayjones/juno/internal/inspection/domain"
)

// PDFGenerator writes an inspection report PDF to the given output path.
type PDFGenerator interface {
	Generate(ctx context.Context, insp *inspectiondomain.Inspection, outputPath string) error
}

// EmailSender delivers the report PDF to a single recipient email address.
// Implementations must return ErrDeliveryQueued when operating in queue-only
// mode so the service knows to leave the delivery in pending state rather than
// marking it failed.
type EmailSender interface {
	SendReport(ctx context.Context, toEmail, pdfPath string) error
}

// ErrDeliveryQueued is returned by queue-only EmailSender implementations to
// signal that the delivery was accepted for later processing — not that an error
// occurred. The service leaves the delivery in "pending" state when it sees this.
var ErrDeliveryQueued = errors.New("delivery queued for later send")
