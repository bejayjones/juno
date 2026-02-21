package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/bejayjones/juno/api/rest"
	inspectionapp "github.com/bejayjones/juno/internal/inspection/application"
	inspectionsqlite "github.com/bejayjones/juno/internal/inspection/infrastructure/sqlite"
	identityapp "github.com/bejayjones/juno/internal/identity/application"
	identityauth "github.com/bejayjones/juno/internal/identity/infrastructure/auth"
	identitysqlite "github.com/bejayjones/juno/internal/identity/infrastructure/sqlite"
	"github.com/bejayjones/juno/internal/platform/db"
	schedulingapp "github.com/bejayjones/juno/internal/scheduling/application"
	schedulingsqlite "github.com/bejayjones/juno/internal/scheduling/infrastructure/sqlite"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/config"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	database, err := db.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		slog.Error("failed to open database", "driver", cfg.Database.Driver, "error", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := database.Migrate(context.Background()); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Identity infrastructure.
	jwtSvc := identityauth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenTTLHours)
	hasher := identityauth.BcryptHasher{}
	companyRepo := identitysqlite.NewCompanyRepository(database)
	inspectorRepo := identitysqlite.NewInspectorRepository(database)
	clientRepo := identitysqlite.NewClientRepository(database)

	// Identity application services.
	clk := clock.Real()
	inspectorSvc := identityapp.NewInspectorService(inspectorRepo, companyRepo, hasher, jwtSvc, clk)
	companySvc := identityapp.NewCompanyService(companyRepo, clk)
	clientSvc := identityapp.NewClientService(clientRepo, clk)

	// Scheduling infrastructure and service.
	appointmentRepo := schedulingsqlite.NewAppointmentRepository(database)
	appointmentSvc := schedulingapp.NewAppointmentService(appointmentRepo, clk)

	// Inspection infrastructure and service.
	inspectionRepo := inspectionsqlite.NewInspectionRepository(database)
	inspectionSvc := inspectionapp.NewInspectionService(inspectionRepo, clk)

	tokenVerifier := rest.NewJWTAdapter(jwtSvc)

	slog.Info("starting juno",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"mode", cfg.Server.Mode,
		"db_driver", cfg.Database.Driver,
	)

	srv := rest.NewServer(cfg, database, inspectorSvc, companySvc, clientSvc, appointmentSvc, inspectionSvc, tokenVerifier)
	if err := srv.Start(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
