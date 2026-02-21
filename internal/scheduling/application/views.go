package application

import "github.com/bejayjones/juno/internal/scheduling/domain"

// PropertyAddressView is the read model for a property address.
type PropertyAddressView struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

// AppointmentView is the read model for an Appointment aggregate.
type AppointmentView struct {
	ID          string              `json:"id"`
	InspectorID string              `json:"inspector_id"`
	ClientID    string              `json:"client_id"`
	Property    PropertyAddressView `json:"property"`
	ScheduledAt int64               `json:"scheduled_at"`
	DurationMin int                 `json:"duration_min"`
	Status      string              `json:"status"`
	Notes       string              `json:"notes"`
	CreatedAt   int64               `json:"created_at"`
	UpdatedAt   int64               `json:"updated_at"`
}

func toAppointmentView(a *domain.Appointment) AppointmentView {
	return AppointmentView{
		ID:          string(a.ID),
		InspectorID: string(a.InspectorID),
		ClientID:    string(a.ClientID),
		Property: PropertyAddressView{
			Street:  a.Property.Street,
			City:    a.Property.City,
			State:   a.Property.State,
			Zip:     a.Property.Zip,
			Country: a.Property.Country,
		},
		ScheduledAt: a.ScheduledAt.Unix(),
		DurationMin: a.EstimatedDurationMin,
		Status:      string(a.Status),
		Notes:       a.Notes,
		CreatedAt:   a.CreatedAt.Unix(),
		UpdatedAt:   a.UpdatedAt.Unix(),
	}
}
