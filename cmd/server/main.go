package main

import (
	"context"
	"log/slog"
	"os"

	"path/filepath"

	"github.com/bejayjones/juno/api/rest"
	inspectionapp "github.com/bejayjones/juno/internal/inspection/application"
	inspectionsqlite "github.com/bejayjones/juno/internal/inspection/infrastructure/sqlite"
	identityapp "github.com/bejayjones/juno/internal/identity/application"
	identityauth "github.com/bejayjones/juno/internal/identity/infrastructure/auth"
	identitysqlite "github.com/bejayjones/juno/internal/identity/infrastructure/sqlite"
	"github.com/bejayjones/juno/internal/platform/db"
	reportingapp "github.com/bejayjones/juno/internal/reporting/application"
	reportingemail "github.com/bejayjones/juno/internal/reporting/infrastructure/email"
	reportingpdf "github.com/bejayjones/juno/internal/reporting/infrastructure/pdf"
	reportingsqlite "github.com/bejayjones/juno/internal/reporting/infrastructure/sqlite"
	schedulingapp "github.com/bejayjones/juno/internal/scheduling/application"
	schedulingsqlite "github.com/bejayjones/juno/internal/scheduling/infrastructure/sqlite"
	syncapp "github.com/bejayjones/juno/internal/sync/application"
	syncdomain "github.com/bejayjones/juno/internal/sync/domain"
	syncsqlite "github.com/bejayjones/juno/internal/sync/infrastructure/sqlite"
	"github.com/bejayjones/juno/internal/sync/recorder"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/config"
	"github.com/bejayjones/juno/pkg/storage"
	"github.com/bejayjones/juno/pkg/storage/local"
	s3storage "github.com/bejayjones/juno/pkg/storage/s3"
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

	// Sync infrastructure: seed the Lamport clock from the stored max value.
	syncRepo := syncsqlite.NewSyncRepository(database)
	maxClock, err := syncRepo.MaxClock(context.Background())
	if err != nil {
		slog.Error("failed to read max lamport clock", "error", err)
		os.Exit(1)
	}
	lamportClock := syncdomain.NewLamportClock(maxClock)
	syncRecorder := recorder.New(lamportClock)
	syncSvc := syncapp.NewSyncService(syncRepo, lamportClock)

	// Identity infrastructure.
	jwtSvc := identityauth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.TokenTTLHours)
	hasher := identityauth.BcryptHasher{}
	companyRepo := identitysqlite.NewCompanyRepository(database).WithRecorder(syncRecorder)
	inspectorRepo := identitysqlite.NewInspectorRepository(database).WithRecorder(syncRecorder)
	clientRepo := identitysqlite.NewClientRepository(database).WithRecorder(syncRecorder)

	// Identity application services.
	clk := clock.Real()
	inspectorSvc := identityapp.NewInspectorService(inspectorRepo, companyRepo, hasher, jwtSvc, clk)
	companySvc := identityapp.NewCompanyService(companyRepo, clk)
	clientSvc := identityapp.NewClientService(clientRepo, clk)

	// Scheduling infrastructure and service.
	appointmentRepo := schedulingsqlite.NewAppointmentRepository(database).WithRecorder(syncRecorder)
	appointmentSvc := schedulingapp.NewAppointmentService(appointmentRepo, clk)

	// Photo storage.
	var photoStore storage.PhotoStorage
	switch cfg.Storage.Driver {
	case "s3":
		photoStore = s3storage.New()
	default:
		photoStore = local.New(cfg.Storage.LocalPath)
	}

	// Inspection infrastructure and service.
	inspectionRepo := inspectionsqlite.NewInspectionRepository(database).WithRecorder(syncRecorder)
	inspectionSvc := inspectionapp.NewInspectionService(inspectionRepo, photoStore, clk)

	// Reporting infrastructure and service.
	reportRepo := reportingsqlite.NewReportRepository(database).WithRecorder(syncRecorder)
	pdfGen := reportingpdf.NewGenerator()
	reportsDir := filepath.Join(filepath.Dir(cfg.Storage.LocalPath), "reports")

	var emailSvc reportingapp.EmailSender
	if cfg.Email.Driver == "smtp" {
		emailSvc = reportingemail.NewSMTPSender(cfg.Email.SMTPHost, cfg.Email.SMTPPort, cfg.Email.SMTPUser, cfg.Email.SMTPPass)
	} else {
		emailSvc = reportingemail.NewQueueOnlySender()
	}
	reportSvc := reportingapp.NewReportService(reportRepo, inspectionRepo, pdfGen, emailSvc, reportsDir, clk)

	tokenVerifier := rest.NewJWTAdapter(jwtSvc)

	slog.Info("starting juno",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"mode", cfg.Server.Mode,
		"db_driver", cfg.Database.Driver,
	)

	srv := rest.NewServer(cfg, database, inspectorSvc, companySvc, clientSvc, appointmentSvc, inspectionSvc, reportSvc, syncSvc, tokenVerifier)
	if err := srv.Start(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
