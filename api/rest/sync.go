package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	syncdomain "github.com/bejayjones/juno/internal/sync/domain"
)

// handleSyncStatus GET /api/v1/sync/status
func (s *Server) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	view, err := s.syncSvc.GetStatus(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "sync status unavailable")
		return
	}
	respond(w, http.StatusOK, view)
}

// handleSyncPush POST /api/v1/sync/push
// Body: {"records": [{id, table_name, record_id, operation, payload, lamport_clock, created_at}, ...]}
func (s *Server) handleSyncPush(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Records []struct {
			ID           string `json:"id"`
			TableName    string `json:"table_name"`
			RecordID     string `json:"record_id"`
			Operation    string `json:"operation"`
			Payload      string `json:"payload"`
			LamportClock int64  `json:"lamport_clock"`
			CreatedAt    int64  `json:"created_at"`
		} `json:"records"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid push body")
		return
	}

	records := make([]*syncdomain.SyncRecord, 0, len(req.Records))
	for _, ri := range req.Records {
		records = append(records, &syncdomain.SyncRecord{
			ID:           ri.ID,
			TableName:    ri.TableName,
			RecordID:     ri.RecordID,
			Operation:    ri.Operation,
			Payload:      ri.Payload,
			LamportClock: ri.LamportClock,
			CreatedAt:    time.Unix(ri.CreatedAt, 0).UTC(),
		})
	}

	result, err := s.syncSvc.HandlePush(r.Context(), records)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "sync push failed")
		return
	}
	respond(w, http.StatusOK, result)
}

// handleSyncPull GET /api/v1/sync/pull?since=<lamport_clock>
func (s *Server) handleSyncPull(w http.ResponseWriter, r *http.Request) {
	since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)

	result, err := s.syncSvc.HandlePull(r.Context(), since)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "sync pull failed")
		return
	}
	respond(w, http.StatusOK, result)
}
