package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bejayjones/juno/api/rest/middleware"
	"github.com/bejayjones/juno/internal/identity/application"
	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/go-chi/chi/v5"
)

// GET /api/v1/me
func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	view, err := s.inspectorSvc.GetByID(r.Context(), domain.InspectorID(p.InspectorID))
	if err != nil {
		if errors.Is(err, domain.ErrInspectorNotFound) {
			respondError(w, http.StatusNotFound, "inspector not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, view)
}

// PUT /api/v1/me
func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())

	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.FirstName == "" || body.LastName == "" || body.Email == "" {
		respondError(w, http.StatusBadRequest, "first_name, last_name, and email are required")
		return
	}

	view, err := s.inspectorSvc.UpdateProfile(r.Context(), domain.InspectorID(p.InspectorID), application.UpdateProfileInput{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInspectorNotFound) {
			respondError(w, http.StatusNotFound, "inspector not found")
			return
		}
		if errors.Is(err, domain.ErrEmailTaken) {
			respondError(w, http.StatusConflict, "email address is already in use")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, view)
}

// PUT /api/v1/me/licenses/{state}
func (s *Server) handleSetLicense(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	state := chi.URLParam(r, "state")

	var body struct {
		Number string `json:"number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Number == "" {
		respondError(w, http.StatusBadRequest, "number is required")
		return
	}

	view, err := s.inspectorSvc.SetLicense(r.Context(), domain.InspectorID(p.InspectorID), application.SetLicenseInput{
		State:  state,
		Number: body.Number,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInspectorNotFound) {
			respondError(w, http.StatusNotFound, "inspector not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, view)
}

// GET /api/v1/companies/{id}/inspectors
func (s *Server) handleListCompanyInspectors(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	p, _ := middleware.PrincipalFromContext(r.Context())
	if p.CompanyID != companyID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	views, err := s.inspectorSvc.ListByCompany(r.Context(), domain.CompanyID(companyID))
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, views)
}
