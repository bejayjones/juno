package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bejayjones/juno/internal/identity/application"
	"github.com/bejayjones/juno/internal/identity/domain"
)

// POST /api/v1/auth/register
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		CompanyID   string `json:"company_id"`
		CompanyName string `json:"company_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.FirstName == "" || body.LastName == "" || body.Email == "" || body.Password == "" {
		respondError(w, http.StatusBadRequest, "first_name, last_name, email, and password are required")
		return
	}

	out, err := s.inspectorSvc.Register(r.Context(), application.RegisterInput{
		FirstName:   body.FirstName,
		LastName:    body.LastName,
		Email:       body.Email,
		Password:    body.Password,
		CompanyID:   body.CompanyID,
		CompanyName: body.CompanyName,
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) {
			respondError(w, http.StatusConflict, "email address is already in use")
			return
		}
		if errors.Is(err, domain.ErrCompanyNotFound) {
			respondError(w, http.StatusUnprocessableEntity, "company not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "registration failed")
		return
	}

	respond(w, http.StatusCreated, out)
}

// POST /api/v1/auth/login
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	out, err := s.inspectorSvc.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInspectorNotFound) {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		respondError(w, http.StatusInternalServerError, "login failed")
		return
	}

	respond(w, http.StatusOK, out)
}
