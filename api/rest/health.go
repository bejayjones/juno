package rest

import (
	"net/http"
	"time"
)

type healthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Mode      string    `json:"mode"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Mode:      s.cfg.Server.Mode,
	})
}
