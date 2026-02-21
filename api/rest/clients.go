package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/bejayjones/juno/api/rest/middleware"
	"github.com/bejayjones/juno/internal/identity/application"
	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/go-chi/chi/v5"
)

// POST /api/v1/clients
func (s *Server) handleCreateClient(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())

	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.FirstName == "" || body.LastName == "" {
		respondError(w, http.StatusBadRequest, "first_name and last_name are required")
		return
	}

	view, err := s.clientSvc.Create(r.Context(), application.CreateClientInput{
		CompanyID: p.CompanyID,
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		Phone:     body.Phone,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusCreated, view)
}

// GET /api/v1/clients
func (s *Server) handleListClients(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	q := r.URL.Query()

	filter := domain.ClientFilter{
		Search: q.Get("search"),
		Limit:  parseIntParam(q.Get("limit"), 50),
		Offset: parseIntParam(q.Get("offset"), 0),
	}

	views, err := s.clientSvc.List(r.Context(), domain.CompanyID(p.CompanyID), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, views)
}

// GET /api/v1/clients/{id}
func (s *Server) handleGetClient(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	clientID := chi.URLParam(r, "id")

	view, err := s.clientSvc.GetByID(r.Context(), domain.ClientID(clientID))
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			respondError(w, http.StatusNotFound, "client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if view.CompanyID != p.CompanyID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}
	respond(w, http.StatusOK, view)
}

// PUT /api/v1/clients/{id}
func (s *Server) handleUpdateClient(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	clientID := chi.URLParam(r, "id")

	existing, err := s.clientSvc.GetByID(r.Context(), domain.ClientID(clientID))
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			respondError(w, http.StatusNotFound, "client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if existing.CompanyID != p.CompanyID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	var body struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.FirstName == "" || body.LastName == "" {
		respondError(w, http.StatusBadRequest, "first_name and last_name are required")
		return
	}

	view, err := s.clientSvc.Update(r.Context(), domain.ClientID(clientID), application.UpdateClientInput{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
		Phone:     body.Phone,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusOK, view)
}

// DELETE /api/v1/clients/{id}
func (s *Server) handleDeleteClient(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	clientID := chi.URLParam(r, "id")

	existing, err := s.clientSvc.GetByID(r.Context(), domain.ClientID(clientID))
	if err != nil {
		if errors.Is(err, domain.ErrClientNotFound) {
			respondError(w, http.StatusNotFound, "client not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if existing.CompanyID != p.CompanyID {
		respondError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := s.clientSvc.Delete(r.Context(), domain.ClientID(clientID)); err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	respond(w, http.StatusNoContent, nil)
}

func parseIntParam(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}
