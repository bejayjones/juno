package rest

import (
	"fmt"
	"net/http"

	"github.com/bejayjones/juno/api/rest/middleware"
	"github.com/bejayjones/juno/pkg/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Server is the HTTP server for the Juno API.
// Application service dependencies are added to this struct as each bounded
// context is implemented in later phases.
type Server struct {
	cfg    *config.Config
	router chi.Router
}

func NewServer(cfg *config.Config) *Server {
	s := &Server{cfg: cfg}
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
