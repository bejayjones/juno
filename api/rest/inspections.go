package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bejayjones/juno/api/rest/middleware"
	inspectionapp "github.com/bejayjones/juno/internal/inspection/application"
	"github.com/bejayjones/juno/internal/inspection/domain"
	"github.com/bejayjones/juno/pkg/storage"
)

// POST /api/v1/inspections — start a new inspection.
func (s *Server) handleStartInspection(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())

	var body struct {
		AppointmentID string   `json:"appointment_id"`
		Weather       string   `json:"weather"`
		TemperatureF  int      `json:"temperature_f"`
		Attendees     []string `json:"attendees"`
		YearBuilt     int      `json:"year_built"`
		StructureType string   `json:"structure_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.AppointmentID == "" {
		respondError(w, http.StatusBadRequest, "appointment_id is required")
		return
	}

	view, err := s.inspectionSvc.Start(r.Context(), inspectionapp.StartInput{
		AppointmentID: body.AppointmentID,
		InspectorID:   p.InspectorID,
		Weather:       body.Weather,
		TemperatureF:  body.TemperatureF,
		Attendees:     body.Attendees,
		YearBuilt:     body.YearBuilt,
		StructureType: body.StructureType,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, view)
}

// GET /api/v1/inspections — list inspections for the authenticated inspector.
func (s *Server) handleListInspections(w http.ResponseWriter, r *http.Request) {
	p, _ := middleware.PrincipalFromContext(r.Context())
	q := r.URL.Query()

	filter := domain.InspectionFilter{
		Limit:  parseIntParam(q.Get("limit"), 50),
		Offset: parseIntParam(q.Get("offset"), 0),
	}
	if statusStr := q.Get("status"); statusStr != "" {
		st := domain.InspectionStatus(statusStr)
		filter.Status = &st
	}
	if fromStr := q.Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.FromDate = &t
		}
	}
	if toStr := q.Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.ToDate = &t
		}
	}

	views, err := s.inspectionSvc.List(r.Context(), p.InspectorID, filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, views)
}

// GET /api/v1/inspections/{id} — get a single inspection.
func (s *Server) handleGetInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	view, err := s.inspectionSvc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}

// GET /api/v1/inspections/{id}/systems — list all system sections.
func (s *Server) handleListSystems(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	view, err := s.inspectionSvc.GetByID(r.Context(), id)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view.Systems)
}

// GET /api/v1/inspections/{id}/systems/{systemType} — get one system section.
func (s *Server) handleGetSystem(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")

	view, err := s.inspectionSvc.GetSystemSection(r.Context(), inspID, systemType)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInvalidSystemType) {
		respondError(w, http.StatusNotFound, "system not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}

// PUT /api/v1/inspections/{id}/systems/{systemType}/descriptions — set "shall describe" fields.
func (s *Server) handleSetDescriptions(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")

	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := s.inspectionSvc.SetDescriptions(r.Context(), inspID, systemType, body)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	if errors.Is(err, domain.ErrInvalidSystemType) {
		respondError(w, http.StatusNotFound, "system not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}

// PUT /api/v1/inspections/{id}/systems/{systemType}/items/{itemKey}/status — set item status.
func (s *Server) handleSetItemStatus(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")
	itemKey := chi.URLParam(r, "itemKey")

	var body struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Status == "" {
		respondError(w, http.StatusBadRequest, "status is required")
		return
	}

	view, err := s.inspectionSvc.SetItemStatus(r.Context(), inspID, systemType, itemKey, body.Status, body.Reason)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	if errors.Is(err, domain.ErrInvalidSystemType) || errors.Is(err, domain.ErrItemNotFound) {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, domain.ErrNIReasonRequired) {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}

// POST /api/v1/inspections/{id}/systems/{systemType}/items/{itemKey}/findings — add a finding.
func (s *Server) handleAddFinding(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")
	itemKey := chi.URLParam(r, "itemKey")

	var body struct {
		Narrative    string `json:"narrative"`
		IsDeficiency bool   `json:"is_deficiency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := s.inspectionSvc.AddFinding(r.Context(), inspID, systemType, itemKey,
		inspectionapp.AddFindingInput{
			Narrative:    body.Narrative,
			IsDeficiency: body.IsDeficiency,
		})
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	if errors.Is(err, domain.ErrInvalidSystemType) || errors.Is(err, domain.ErrItemNotFound) {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, view)
}

// PUT /api/v1/inspections/{id}/systems/{systemType}/items/{itemKey}/findings/{findingID} — update a finding.
func (s *Server) handleUpdateFinding(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")
	itemKey := chi.URLParam(r, "itemKey")
	findingID := chi.URLParam(r, "findingID")

	var body struct {
		Narrative    string `json:"narrative"`
		IsDeficiency bool   `json:"is_deficiency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := s.inspectionSvc.UpdateFinding(r.Context(), inspID, systemType, itemKey, findingID,
		inspectionapp.UpdateFindingInput{
			Narrative:    body.Narrative,
			IsDeficiency: body.IsDeficiency,
		})
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	if errors.Is(err, domain.ErrInvalidSystemType) || errors.Is(err, domain.ErrItemNotFound) || errors.Is(err, domain.ErrFindingNotFound) {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}

// DELETE /api/v1/inspections/{id}/systems/{systemType}/items/{itemKey}/findings/{findingID} — remove a finding.
func (s *Server) handleDeleteFinding(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")
	itemKey := chi.URLParam(r, "itemKey")
	findingID := chi.URLParam(r, "findingID")

	err := s.inspectionSvc.DeleteFinding(r.Context(), inspID, systemType, itemKey, findingID)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	if errors.Is(err, domain.ErrInvalidSystemType) || errors.Is(err, domain.ErrItemNotFound) || errors.Is(err, domain.ErrFindingNotFound) {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/inspections/{id}/complete — finalize an inspection.
func (s *Server) handleCompleteInspection(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")

	view, err := s.inspectionSvc.Complete(r.Context(), inspID)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	// ValidationError: missing items/descriptions.
	var ve *domain.ValidationError
	if errors.As(err, &ve) {
		respond(w, http.StatusUnprocessableEntity, map[string]any{
			"error":  "inspection is not ready to complete",
			"fields": ve.Fields,
		})
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, view)
}

// GET /api/v1/inspections/{id}/summary — deficiency summary.
func (s *Server) handleGetDeficiencySummary(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")

	views, err := s.inspectionSvc.GetDeficiencySummary(r.Context(), inspID)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusOK, views)
}

// POST /api/v1/inspections/{id}/systems/{systemType}/items/{itemKey}/photos
// Accepts multipart/form-data with fields: finding_id (text), photo (file).
func (s *Server) handleAddPhoto(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")
	itemKey := chi.URLParam(r, "itemKey")

	// Enforce 20 MB ceiling on the entire multipart body.
	r.Body = http.MaxBytesReader(w, r.Body, storage.MaxPhotoBytes+1<<20)
	if err := r.ParseMultipartForm(storage.MaxPhotoBytes); err != nil {
		respondError(w, http.StatusRequestEntityTooLarge, "photo exceeds 20 MB limit")
		return
	}

	findingID := r.FormValue("finding_id")
	if findingID == "" {
		respondError(w, http.StatusBadRequest, "finding_id is required")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		respondError(w, http.StatusBadRequest, "photo file is required")
		return
	}
	defer file.Close()

	// Determine MIME type: prefer explicit Content-Type, fall back to filename extension.
	mimeType := header.Header.Get("Content-Type")
	if _, ok := storage.AllowedMimeTypes[mimeType]; !ok {
		respondError(w, http.StatusUnsupportedMediaType, domain.ErrInvalidMimeType.Error())
		return
	}

	view, err := s.inspectionSvc.AddPhoto(r.Context(),
		inspID, systemType, itemKey, findingID, mimeType, file)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	if errors.Is(err, domain.ErrInvalidSystemType) || errors.Is(err, domain.ErrItemNotFound) || errors.Is(err, domain.ErrFindingNotFound) {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, domain.ErrInvalidMimeType) {
		respondError(w, http.StatusUnsupportedMediaType, err.Error())
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respond(w, http.StatusCreated, view)
}

// DELETE /api/v1/inspections/{id}/systems/{systemType}/items/{itemKey}/photos/{photoID}
func (s *Server) handleDeletePhoto(w http.ResponseWriter, r *http.Request) {
	inspID := chi.URLParam(r, "id")
	systemType := chi.URLParam(r, "systemType")
	itemKey := chi.URLParam(r, "itemKey")
	photoID := chi.URLParam(r, "photoID")

	err := s.inspectionSvc.DeletePhoto(r.Context(), inspID, systemType, itemKey, photoID)
	if errors.Is(err, domain.ErrInspectionNotFound) {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	if errors.Is(err, domain.ErrInspectionCompleted) {
		respondError(w, http.StatusUnprocessableEntity, "inspection is already completed")
		return
	}
	if errors.Is(err, domain.ErrPhotoNotFound) || errors.Is(err, domain.ErrInvalidSystemType) || errors.Is(err, domain.ErrItemNotFound) {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/photos/{photoID} — stream a photo file.
func (s *Server) handleServePhoto(w http.ResponseWriter, r *http.Request) {
	photoID := chi.URLParam(r, "photoID")

	rc, mimeType, err := s.inspectionSvc.GetPhotoData(r.Context(), photoID)
	if errors.Is(err, domain.ErrPhotoNotFound) {
		respondError(w, http.StatusNotFound, "photo not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rc.Close()

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "private, max-age=3600")
	io.Copy(w, rc) //nolint:errcheck
}
