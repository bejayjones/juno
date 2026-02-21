package rest

import "github.com/go-chi/chi/v5"

func (s *Server) registerRoutes(r chi.Router) {
	r.Get("/health", s.handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		// Identity, scheduling, inspection, and reporting routes
		// are registered here in later phases.
	})
}
