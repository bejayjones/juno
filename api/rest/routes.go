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

			// Scheduling.
			r.Post("/appointments", s.handleCreateAppointment)
			r.Get("/appointments", s.handleListAppointments)
			r.Get("/appointments/{id}", s.handleGetAppointment)
			r.Put("/appointments/{id}", s.handleUpdateAppointment)
			r.Delete("/appointments/{id}", s.handleCancelAppointment)

			// Inspection walkthrough.
			r.Post("/inspections", s.handleStartInspection)
			r.Get("/inspections", s.handleListInspections)
			r.Get("/inspections/{id}", s.handleGetInspection)
			r.Post("/inspections/{id}/complete", s.handleCompleteInspection)
			r.Get("/inspections/{id}/summary", s.handleGetDeficiencySummary)
			r.Get("/inspections/{id}/systems", s.handleListSystems)
			r.Get("/inspections/{id}/systems/{systemType}", s.handleGetSystem)
			r.Put("/inspections/{id}/systems/{systemType}/descriptions", s.handleSetDescriptions)
			r.Put("/inspections/{id}/systems/{systemType}/items/{itemKey}/status", s.handleSetItemStatus)
			r.Post("/inspections/{id}/systems/{systemType}/items/{itemKey}/findings", s.handleAddFinding)
			r.Put("/inspections/{id}/systems/{systemType}/items/{itemKey}/findings/{findingID}", s.handleUpdateFinding)
			r.Delete("/inspections/{id}/systems/{systemType}/items/{itemKey}/findings/{findingID}", s.handleDeleteFinding)

			// Photo upload/delete (authenticated; serve is also public below).
			r.Post("/inspections/{id}/systems/{systemType}/items/{itemKey}/photos", s.handleAddPhoto)
			r.Delete("/inspections/{id}/systems/{systemType}/items/{itemKey}/photos/{photoID}", s.handleDeletePhoto)

			// Photo streaming (inside auth group so the token is required).
			r.Get("/photos/{photoID}", s.handleServePhoto)

			// Reporting.
			r.Post("/reports", s.handleGenerateReport)
			r.Get("/reports", s.handleListReports)
			r.Get("/reports/{id}", s.handleGetReport)
			r.Get("/reports/{id}/pdf", s.handleGetReportPDF)
			r.Put("/reports/{id}/finalize", s.handleFinalizeReport)
			r.Post("/reports/{id}/deliver", s.handleDeliverReport)
			r.Get("/reports/{id}/deliveries", s.handleListDeliveries)
			r.Post("/reports/{id}/deliveries/retry", s.handleRetryDeliveries)

			// Sync.
			r.Get("/sync/status", s.handleSyncStatus)
			r.Post("/sync/push", s.handleSyncPush)
			r.Get("/sync/pull", s.handleSyncPull)
		})
	})

	// Serve the SvelteKit SPA for all other paths.
	r.Handle("/*", spaHandler())
}
