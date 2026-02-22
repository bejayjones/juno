package application

import "github.com/bejayjones/juno/internal/reporting/domain"

// ReportView is the read model returned to REST clients.
type ReportView struct {
	ID             string         `json:"id"`
	InspectionID   string         `json:"inspection_id"`
	InspectorID    string         `json:"inspector_id"`
	Status         string         `json:"status"`
	PDFStoragePath string         `json:"pdf_storage_path,omitempty"`
	GeneratedAt    *int64         `json:"generated_at,omitempty"`
	Deliveries     []DeliveryView `json:"deliveries"`
	CreatedAt      int64          `json:"created_at"`
	UpdatedAt      int64          `json:"updated_at"`
}

// DeliveryView is the read model for a single email delivery attempt.
type DeliveryView struct {
	ID             string `json:"id"`
	RecipientEmail string `json:"recipient_email"`
	Status         string `json:"status"`
	Attempts       int    `json:"attempts"`
	SentAt         *int64 `json:"sent_at,omitempty"`
	FailureReason  string `json:"failure_reason,omitempty"`
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

func toReportView(r *domain.Report) ReportView {
	v := ReportView{
		ID:             string(r.ID),
		InspectionID:   string(r.InspectionID),
		InspectorID:    string(r.InspectorID),
		Status:         string(r.Status),
		PDFStoragePath: r.PDFStoragePath,
		Deliveries:     make([]DeliveryView, len(r.Deliveries)),
		CreatedAt:      r.CreatedAt.Unix(),
		UpdatedAt:      r.UpdatedAt.Unix(),
	}
	if r.GeneratedAt != nil {
		t := r.GeneratedAt.Unix()
		v.GeneratedAt = &t
	}
	for i, d := range r.Deliveries {
		v.Deliveries[i] = toDeliveryView(d)
	}
	return v
}

func toDeliveryView(d domain.Delivery) DeliveryView {
	v := DeliveryView{
		ID:             string(d.ID),
		RecipientEmail: d.RecipientEmail,
		Status:         string(d.Status),
		Attempts:       d.Attempts,
		FailureReason:  d.FailureReason,
		CreatedAt:      d.CreatedAt.Unix(),
		UpdatedAt:      d.UpdatedAt.Unix(),
	}
	if d.SentAt != nil {
		t := d.SentAt.Unix()
		v.SentAt = &t
	}
	return v
}
