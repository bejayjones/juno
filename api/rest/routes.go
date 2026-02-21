package rest

import (
	"github.com/bejayjones/juno/api/rest/middleware"
	"github.com/go-chi/chi/v5"
)

func (s *Server) registerRoutes(r chi.Router) {
	r.Get("/health", s.handleHealth)

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth endpoints.
		r.Post("/auth/register", s.handleRegister)
		r.Post("/auth/login", s.handleLogin)

		// Authenticated endpoints.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(s.tokenVerifier))

			// Inspector profile.
			r.Get("/me", s.handleGetMe)
			r.Put("/me", s.handleUpdateMe)
			r.Put("/me/licenses/{state}", s.handleSetLicense)

			// Company management.
			r.Get("/companies/{id}", s.handleGetCompany)
			r.Put("/companies/{id}", s.handleUpdateCompany)
			r.Get("/companies/{id}/inspectors", s.handleListCompanyInspectors)

			// Client roster.
			r.Post("/clients", s.handleCreateClient)
			r.Get("/clients", s.handleListClients)
			r.Get("/clients/{id}", s.handleGetClient)
			r.Put("/clients/{id}", s.handleUpdateClient)
			r.Delete("/clients/{id}", s.handleDeleteClient)
		})
	})
}
