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

// GET /api/v1/companies/{id}
func (s *Server) handleGetCompany(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	p, _ := middleware.PrincipalFromContext(r.Context())
	if p.CompanyID != companyID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	view, err := s.companySvc.GetByID(r.Context(), domain.CompanyID(companyID))
	if err != nil {
		if errors.Is(err, domain.ErrCompanyNotFound) {
			respondError(w, http.StatusNotFound, "company not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, view)
}

// PUT /api/v1/companies/{id}
// Restricted to the company owner.
func (s *Server) handleUpdateCompany(w http.ResponseWriter, r *http.Request) {
	companyID := chi.URLParam(r, "id")
	p, _ := middleware.PrincipalFromContext(r.Context())
	if p.CompanyID != companyID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}
	if p.Role != string(domain.RoleOwner) {
		respondError(w, http.StatusForbidden, "only the company owner can update company details")
		return
	}

	var body struct {
		Name    string `json:"name"`
		Street  string `json:"street"`
		City    string `json:"city"`
		State   string `json:"state"`
		Zip     string `json:"zip"`
		Country string `json:"country"`
		Phone   string `json:"phone"`
		Email   string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	view, err := s.companySvc.Update(r.Context(), domain.CompanyID(companyID), application.UpdateCompanyInput{
		Name:    body.Name,
		Street:  body.Street,
		City:    body.City,
		State:   body.State,
		Zip:     body.Zip,
		Country: body.Country,
		Phone:   body.Phone,
		Email:   body.Email,
	})
	if err != nil {
		if errors.Is(err, domain.ErrCompanyNotFound) {
			respondError(w, http.StatusNotFound, "company not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, view)
}
