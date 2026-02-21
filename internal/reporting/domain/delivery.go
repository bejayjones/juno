package domain

import "time"

type DeliveryID string

type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliverySent    DeliveryStatus = "sent"
	DeliveryFailed_ DeliveryStatus = "failed"
)

// Delivery is an entity within a Report representing one email send attempt
// to a single recipient. Multiple recipients produce multiple Deliveries.
type Delivery struct {
	ID             DeliveryID
	RecipientEmail string
	Status         DeliveryStatus
	Attempts       int
	SentAt         *time.Time
	FailureReason  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewDelivery(id DeliveryID, recipientEmail string, now time.Time) Delivery {
	return Delivery{
		ID:             id,
		RecipientEmail: recipientEmail,
		Status:         DeliveryPending,
		Attempts:       0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (d *Delivery) MarkSent(now time.Time) {
	t := now
	d.SentAt = &t
	d.Status = DeliverySent
	d.Attempts++
	d.UpdatedAt = now
}

func (d *Delivery) MarkFailed(reason string, now time.Time) {
	d.Status = DeliveryFailed_
	d.FailureReason = reason
	d.Attempts++
	d.UpdatedAt = now
}

func (d *Delivery) ResetForRetry(now time.Time) {
	d.Status = DeliveryPending
	d.FailureReason = ""
	d.UpdatedAt = now
}
