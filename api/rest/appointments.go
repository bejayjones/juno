package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/bejayjones/juno/api/rest/middleware"
	"github.com/bejayjones/juno/internal/scheduling/application"
	"github.com/bejayjones/juno/internal/scheduling/domain"
	"github.com/go-chi/chi/v5"
)

// GET /api/v1/appointments
func (s *Server) handleListAppointments(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	q := r.URL.Query()

	in := application.ListInput{
		InspectorID: p.InspectorID,
		Limit:       parseIntParam(q.Get("limit"), 50),
		Offset:      parseIntParam(q.Get("offset"), 0),
	}

	if raw := q.Get("status"); raw != "" {
		st := domain.AppointmentStatus(raw)
		in.Status = &st
	}
	if raw := q.Get("from"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			in.FromDate = &t
		}
	}
	if raw := q.Get("to"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			in.ToDate = &t
		}
	}

	views, err := s.appointmentSvc.List(r.Context(), in)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, views)
}

// POST /api/v1/appointments
func (s *Server) handleCreateAppointment(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())

	var body struct {
		ClientID    string `json:"client_id"`
		Street      string `json:"street"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		ScheduledAt int64  `json:"scheduled_at"` // Unix timestamp
		DurationMin int    `json:"duration_min"`
		Notes       string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.ClientID == "" || body.Street == "" || body.City == "" || body.ScheduledAt == 0 {
		respondError(w, http.StatusBadRequest, "client_id, street, city, and scheduled_at are required")
		return
	}

	view, err := s.appointmentSvc.Schedule(r.Context(), application.ScheduleInput{
		InspectorID: p.InspectorID,
		ClientID:    body.ClientID,
		Street:      body.Street,
		City:        body.City,
		State:       body.State,
		Zip:         body.Zip,
		Country:     body.Country,
		ScheduledAt: time.Unix(body.ScheduledAt, 0).UTC(),
		DurationMin: body.DurationMin,
		Notes:       body.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPastScheduledTime) {
			respondError(w, http.StatusUnprocessableEntity, "scheduled_at must be in the future")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusCreated, view)
}

// GET /api/v1/appointments/{id}
func (s *Server) handleGetAppointment(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	apptID := chi.URLParam(r, "id")

	view, err := s.appointmentSvc.GetByID(r.Context(), domain.AppointmentID(apptID))
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			respondError(w, http.StatusNotFound, "appointment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if view.InspectorID != p.InspectorID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}
	respond(w, http.StatusOK, view)
}

// PUT /api/v1/appointments/{id}
func (s *Server) handleUpdateAppointment(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	apptID := chi.URLParam(r, "id")

	existing, err := s.appointmentSvc.GetByID(r.Context(), domain.AppointmentID(apptID))
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			respondError(w, http.StatusNotFound, "appointment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if existing.InspectorID != p.InspectorID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	var body struct {
		Street      string `json:"street"`
		City        string `json:"city"`
		State       string `json:"state"`
		Zip         string `json:"zip"`
		Country     string `json:"country"`
		ScheduledAt *int64 `json:"scheduled_at"` // nil = no change
		DurationMin int    `json:"duration_min"`
		Notes       string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	in := application.UpdateInput{
		Street:      body.Street,
		City:        body.City,
		State:       body.State,
		Zip:         body.Zip,
		Country:     body.Country,
		DurationMin: body.DurationMin,
		Notes:       body.Notes,
	}
	if body.ScheduledAt != nil {
		t := time.Unix(*body.ScheduledAt, 0).UTC()
		in.ScheduledAt = &t
	}

	view, err := s.appointmentSvc.Update(r.Context(), domain.AppointmentID(apptID), in)
	if err != nil {
		if errors.Is(err, domain.ErrPastScheduledTime) {
			respondError(w, http.StatusUnprocessableEntity, "scheduled_at must be in the future")
			return
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			respondError(w, http.StatusConflict, "appointment cannot be modified in its current status")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, view)
}

// DELETE /api/v1/appointments/{id}
// Cancels the appointment (sets status to cancelled). Does not hard-delete.
func (s *Server) handleCancelAppointment(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	apptID := chi.URLParam(r, "id")

	existing, err := s.appointmentSvc.GetByID(r.Context(), domain.AppointmentID(apptID))
	if err != nil {
		if errors.Is(err, domain.ErrAppointmentNotFound) {
			respondError(w, http.StatusNotFound, "appointment not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if existing.InspectorID != p.InspectorID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := s.appointmentSvc.Cancel(r.Context(), domain.AppointmentID(apptID)); err != nil {
		if errors.Is(err, domain.ErrCannotCancelCompleted) {
			respondError(w, http.StatusConflict, "cannot cancel a completed appointment")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusNoContent, nil)
}
