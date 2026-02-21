package rest

import (
	"fmt"
	"net/http"

	"github.com/bejayjones/juno/api/rest/middleware"
	"github.com/bejayjones/juno/internal/platform/db"
	"github.com/bejayjones/juno/pkg/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Server is the HTTP server for the Juno API.
// Bounded-context application services are added as fields here as each
// phase is implemented.
type Server struct {
	cfg    *config.Config
	db     *db.DB
	router chi.Router
}

func NewServer(cfg *config.Config, database *db.DB) *Server {
	s := &Server{cfg: cfg, db: database}
	s.router = s.buildRouter()
	return s
}

func (s *Server) buildRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)
	s.registerRoutes(r)
	return r
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Start binds to the configured address and blocks serving requests.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	return http.ListenAndServe(addr, s)
}
