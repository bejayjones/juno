package rest

import (
	"fmt"
	"net/http"

	"github.com/bejayjones/juno/api/rest/middleware"
	identityapp "github.com/bejayjones/juno/internal/identity/application"
	identityauth "github.com/bejayjones/juno/internal/identity/infrastructure/auth"
	"github.com/bejayjones/juno/internal/platform/db"
	schedulingapp "github.com/bejayjones/juno/internal/scheduling/application"
	"github.com/bejayjones/juno/pkg/config"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Server is the HTTP server for the Juno API.
type Server struct {
	cfg            *config.Config
	db             *db.DB
	router         chi.Router
	inspectorSvc   *identityapp.InspectorService
	companySvc     *identityapp.CompanyService
	clientSvc      *identityapp.ClientService
	appointmentSvc *schedulingapp.AppointmentService
	tokenVerifier  middleware.TokenVerifier
}

// NewServer wires up the server with all application services.
func NewServer(
	cfg *config.Config,
	database *db.DB,
	inspectorSvc *identityapp.InspectorService,
	companySvc *identityapp.CompanyService,
	clientSvc *identityapp.ClientService,
	appointmentSvc *schedulingapp.AppointmentService,
	tokenVerifier middleware.TokenVerifier,
) *Server {
	s := &Server{
		cfg:            cfg,
		db:             database,
		inspectorSvc:   inspectorSvc,
		companySvc:     companySvc,
		clientSvc:      clientSvc,
		appointmentSvc: appointmentSvc,
		tokenVerifier:  tokenVerifier,
	}
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

// jwtAdapter adapts JWTService.Verify to the middleware.TokenVerifier interface.
type jwtAdapter struct {
	svc *identityauth.JWTService
}

func NewJWTAdapter(svc *identityauth.JWTService) middleware.TokenVerifier {
	return &jwtAdapter{svc: svc}
}

func (a *jwtAdapter) VerifyToken(token string) (middleware.Principal, error) {
	claims, err := a.svc.Verify(token)
	if err != nil {
		return middleware.Principal{}, err
	}
	return middleware.Principal{
		InspectorID: claims.InspectorID,
		CompanyID:   claims.CompanyID,
		Role:        claims.Role,
	}, nil
}
