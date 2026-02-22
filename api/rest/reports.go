package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bejayjones/juno/api/rest/middleware"
	"github.com/bejayjones/juno/internal/reporting/domain"
)

// POST /api/v1/reports — generate a report PDF for a completed inspection.
func (s *Server) handleGenerateReport(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())

	var body struct {
		InspectionID string `json:"inspection_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.InspectionID == "" {
		respondError(w, http.StatusBadRequest, "inspection_id is required")
		return
	}

	view, err := s.reportSvc.GenerateReport(r.Context(), body.InspectionID, p.InspectorID)
	switch {
	case errors.Is(err, domain.ErrReportAlreadyExists):
		respondError(w, http.StatusConflict, "a report already exists for this inspection")
	case errors.Is(err, domain.ErrReportNotFound):
		respondError(w, http.StatusNotFound, "report not found")
	case err != nil:
		respondError(w, http.StatusInternalServerError, err.Error())
	default:
		respond(w, http.StatusCreated, view)
	}
}

// GET /api/v1/reports — list reports for the authenticated inspector.
func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	q := r.URL.Query()

	limit := parseIntParam(q.Get("limit"), 50)
	offset := parseIntParam(q.Get("offset"), 0)

	views, err := s.reportSvc.List(r.Context(), p.InspectorID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, views)
}

// GET /api/v1/reports/{id} — retrieve a single report.
func (s *Server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	view, err := s.reportSvc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrReportNotFound) {
		respondError(w, http.StatusNotFound, "report not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}

// GET /api/v1/reports/{id}/pdf — stream the report PDF file.
func (s *Server) handleGetReportPDF(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	pdfPath, err := s.reportSvc.PDFPath(r.Context(), id)
	switch {
	case errors.Is(err, domain.ErrReportNotFound):
		respondError(w, http.StatusNotFound, "report not found")
	case errors.Is(err, domain.ErrReportNotGenerated):
		respondError(w, http.StatusConflict, "report PDF has not been generated yet")
	case err != nil:
		respondError(w, http.StatusInternalServerError, err.Error())
	default:
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", `attachment; filename="inspection-report.pdf"`)
		http.ServeFile(w, r, pdfPath)
	}
}

// PUT /api/v1/reports/{id}/finalize — lock the report.
func (s *Server) handleFinalizeReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	view, err := s.reportSvc.FinalizeReport(r.Context(), id)
	switch {
	case errors.Is(err, domain.ErrReportNotFound):
		respondError(w, http.StatusNotFound, "report not found")
	case errors.Is(err, domain.ErrReportFinalized):
		respondError(w, http.StatusConflict, "report is already finalized")
	case errors.Is(err, domain.ErrReportNotGenerated):
		respondError(w, http.StatusConflict, "report PDF has not been generated yet")
	case err != nil:
		respondError(w, http.StatusInternalServerError, err.Error())
	default:
		respond(w, http.StatusOK, view)
	}
}

// POST /api/v1/reports/{id}/deliver — queue an email delivery to a recipient.
func (s *Server) handleDeliverReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var body struct {
		RecipientEmail string `json:"recipient_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.RecipientEmail == "" {
		respondError(w, http.StatusBadRequest, "recipient_email is required")
		return
	}

	dv, err := s.reportSvc.QueueDelivery(r.Context(), id, body.RecipientEmail)
	switch {
	case errors.Is(err, domain.ErrReportNotFound):
		respondError(w, http.StatusNotFound, "report not found")
	case errors.Is(err, domain.ErrReportNotGenerated):
		respondError(w, http.StatusConflict, "report PDF has not been generated yet")
	case err != nil:
		respondError(w, http.StatusInternalServerError, err.Error())
	default:
		respond(w, http.StatusCreated, dv)
	}
}

// GET /api/v1/reports/{id}/deliveries — list all deliveries for a report.
func (s *Server) handleListDeliveries(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	view, err := s.reportSvc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrReportNotFound) {
		respondError(w, http.StatusNotFound, "report not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view.Deliveries)
}

// POST /api/v1/reports/{id}/deliveries/retry — retry all failed deliveries.
func (s *Server) handleRetryDeliveries(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	view, err := s.reportSvc.RetryFailedDeliveries(r.Context(), id)
	if errors.Is(err, domain.ErrReportNotFound) {
		respondError(w, http.StatusNotFound, "report not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}
