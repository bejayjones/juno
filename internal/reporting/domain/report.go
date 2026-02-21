package domain

import "time"

type ReportID string
type InspectionID string
type InspectorID string

type ReportStatus string

const (
	ReportDraft     ReportStatus = "draft"
	ReportFinalized ReportStatus = "finalized"
)

// Report is the aggregate root for the reporting context. It is created after
// an Inspection is completed and owns the PDF generation lifecycle and all
// email delivery records.
type Report struct {
	ID             ReportID
	InspectionID   InspectionID
	InspectorID    InspectorID
	Status         ReportStatus
	PDFStoragePath string    // empty until MarkGenerated is called
	GeneratedAt    *time.Time
	Deliveries     []Delivery
	CreatedAt      time.Time
	UpdatedAt      time.Time
	events         []any
}

func NewReport(
	id ReportID,
	inspectionID InspectionID,
	inspectorID InspectorID,
	now time.Time,
) *Report {
	return &Report{
		ID:           id,
		InspectionID: inspectionID,
		InspectorID:  inspectorID,
		Status:       ReportDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// MarkGenerated records the storage path of the generated PDF.
func (r *Report) MarkGenerated(pdfStoragePath string, now time.Time) {
	r.PDFStoragePath = pdfStoragePath
	t := now
	r.GeneratedAt = &t
	r.UpdatedAt = now
	r.record(ReportGenerated{
		ReportID:     r.ID,
		InspectionID: r.InspectionID,
		OccurredAt:   now,
	})
}

// Finalize locks the report. Finalized reports cannot have items edited.
func (r *Report) Finalize(now time.Time) error {
	if r.Status == ReportFinalized {
		return ErrReportFinalized
	}
	if r.PDFStoragePath == "" {
		return ErrReportNotGenerated
	}
	r.Status = ReportFinalized
	r.UpdatedAt = now
	r.record(ReportWasFinalized{ReportID: r.ID, OccurredAt: now})
	return nil
}

// AddDelivery appends a new pending delivery for the given recipient.
func (r *Report) AddDelivery(delivery Delivery, now time.Time) error {
	if r.PDFStoragePath == "" {
		return ErrReportNotGenerated
	}
	r.Deliveries = append(r.Deliveries, delivery)
	r.UpdatedAt = now
	r.record(DeliveryQueued{
		ReportID:       r.ID,
		DeliveryID:     delivery.ID,
		RecipientEmail: delivery.RecipientEmail,
		OccurredAt:     now,
	})
	return nil
}

// MarkDelivered marks a delivery as successfully sent.
func (r *Report) MarkDelivered(deliveryID DeliveryID, now time.Time) error {
	d, err := r.deliveryByID(deliveryID)
	if err != nil {
		return err
	}
	d.MarkSent(now)
	r.UpdatedAt = now
	r.record(DeliverySucceeded{
		ReportID:   r.ID,
		DeliveryID: deliveryID,
		OccurredAt: now,
	})
	return nil
}

// MarkDeliveryFailed records a send failure with the reason.
func (r *Report) MarkDeliveryFailed(deliveryID DeliveryID, reason string, now time.Time) error {
	d, err := r.deliveryByID(deliveryID)
	if err != nil {
		return err
	}
	d.MarkFailed(reason, now)
	r.UpdatedAt = now
	r.record(DeliveryFailed{
		ReportID:      r.ID,
		DeliveryID:    deliveryID,
		FailureReason: reason,
		OccurredAt:    now,
	})
	return nil
}

// RetryDelivery resets a failed delivery back to pending for re-sending.
func (r *Report) RetryDelivery(deliveryID DeliveryID, now time.Time) error {
	d, err := r.deliveryByID(deliveryID)
	if err != nil {
		return err
	}
	d.ResetForRetry(now)
	r.UpdatedAt = now
	return nil
}

func (r *Report) deliveryByID(id DeliveryID) (*Delivery, error) {
	for i := range r.Deliveries {
		if r.Deliveries[i].ID == id {
			return &r.Deliveries[i], nil
		}
	}
	return nil, ErrDeliveryNotFound
}

func (r *Report) Events() []any { return r.events }
func (r *Report) ClearEvents()  { r.events = nil }
func (r *Report) record(e any)  { r.events = append(r.events, e) }
