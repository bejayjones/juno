// Package application implements the reporting use-cases.
package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	inspectiondomain "github.com/bejayjones/juno/internal/inspection/domain"
	"github.com/bejayjones/juno/internal/reporting/domain"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/id"
)

// ReportService orchestrates report generation, finalization, and email delivery.
type ReportService struct {
	reports     domain.ReportRepository
	inspections inspectiondomain.InspectionRepository
	pdfGen      PDFGenerator
	emailSvc    EmailSender
	reportsDir  string
	clk         clock.Clock
}

func NewReportService(
	reports domain.ReportRepository,
	inspections inspectiondomain.InspectionRepository,
	pdfGen PDFGenerator,
	emailSvc EmailSender,
	reportsDir string,
	clk clock.Clock,
) *ReportService {
	return &ReportService{
		reports:     reports,
		inspections: inspections,
		pdfGen:      pdfGen,
		emailSvc:    emailSvc,
		reportsDir:  reportsDir,
		clk:         clk,
	}
}

// GenerateReport creates a draft Report, generates the PDF, and persists both.
// Returns ErrReportAlreadyExists if a report for this inspection already exists.
func (s *ReportService) GenerateReport(ctx context.Context, inspectionID, inspectorID string) (ReportView, error) {
	// Guard: one report per inspection.
	existing, err := s.reports.FindByInspection(ctx, domain.InspectionID(inspectionID))
	if err != nil && !errors.Is(err, domain.ErrReportNotFound) {
		return ReportView{}, err
	}
	if existing != nil {
		return ReportView{}, domain.ErrReportAlreadyExists
	}

	insp, err := s.inspections.FindByID(ctx, inspectiondomain.InspectionID(inspectionID))
	if err != nil {
		return ReportView{}, fmt.Errorf("load inspection: %w", err)
	}

	now := s.clk.Now()
	report := domain.NewReport(
		domain.ReportID(id.New()),
		domain.InspectionID(inspectionID),
		domain.InspectorID(inspectorID),
		now,
	)

	if err := os.MkdirAll(s.reportsDir, 0750); err != nil {
		return ReportView{}, fmt.Errorf("create reports dir: %w", err)
	}

	pdfPath := filepath.Join(s.reportsDir, string(report.ID)+".pdf")
	if err := s.pdfGen.Generate(ctx, insp, pdfPath); err != nil {
		return ReportView{}, fmt.Errorf("generate pdf: %w", err)
	}

	report.MarkGenerated(pdfPath, s.clk.Now())

	if err := s.reports.Save(ctx, report); err != nil {
		_ = os.Remove(pdfPath) // best-effort cleanup of orphaned file
		return ReportView{}, err
	}
	return toReportView(report), nil
}

// FinalizeReport locks the report; finalized reports cannot be regenerated.
func (s *ReportService) FinalizeReport(ctx context.Context, reportID string) (ReportView, error) {
	report, err := s.reports.FindByID(ctx, domain.ReportID(reportID))
	if err != nil {
		return ReportView{}, err
	}
	if err := report.Finalize(s.clk.Now()); err != nil {
		return ReportView{}, err
	}
	if err := s.reports.Save(ctx, report); err != nil {
		return ReportView{}, err
	}
	return toReportView(report), nil
}

// QueueDelivery adds a pending delivery and immediately attempts to send it.
// In queue-only mode the EmailSender returns ErrDeliveryQueued and the delivery
// stays pending; in SMTP mode success/failure are reflected in the delivery status.
func (s *ReportService) QueueDelivery(ctx context.Context, reportID, recipientEmail string) (DeliveryView, error) {
	report, err := s.reports.FindByID(ctx, domain.ReportID(reportID))
	if err != nil {
		return DeliveryView{}, err
	}

	now := s.clk.Now()
	delivery := domain.NewDelivery(domain.DeliveryID(id.New()), recipientEmail, now)
	if err := report.AddDelivery(delivery, now); err != nil {
		return DeliveryView{}, err
	}

	// Attempt immediate send; interpret ErrDeliveryQueued as "leave pending".
	sendErr := s.emailSvc.SendReport(ctx, recipientEmail, report.PDFStoragePath)
	switch {
	case sendErr == nil:
		_ = report.MarkDelivered(delivery.ID, s.clk.Now())
	case errors.Is(sendErr, ErrDeliveryQueued):
		// leave delivery in pending state
	default:
		_ = report.MarkDeliveryFailed(delivery.ID, sendErr.Error(), s.clk.Now())
	}

	if err := s.reports.Save(ctx, report); err != nil {
		return DeliveryView{}, err
	}

	// Return the updated delivery.
	for _, d := range report.Deliveries {
		if d.ID == delivery.ID {
			return toDeliveryView(d), nil
		}
	}
	return toDeliveryView(delivery), nil
}

// RetryFailedDeliveries resets failed deliveries to pending and retries sending.
func (s *ReportService) RetryFailedDeliveries(ctx context.Context, reportID string) (ReportView, error) {
	report, err := s.reports.FindByID(ctx, domain.ReportID(reportID))
	if err != nil {
		return ReportView{}, err
	}

	for _, d := range report.Deliveries {
		if d.Status != domain.DeliveryFailed_ {
			continue
		}
		_ = report.RetryDelivery(d.ID, s.clk.Now())
		sendErr := s.emailSvc.SendReport(ctx, d.RecipientEmail, report.PDFStoragePath)
		switch {
		case sendErr == nil:
			_ = report.MarkDelivered(d.ID, s.clk.Now())
		case errors.Is(sendErr, ErrDeliveryQueued):
			// leave as pending
		default:
			_ = report.MarkDeliveryFailed(d.ID, sendErr.Error(), s.clk.Now())
		}
	}

	if err := s.reports.Save(ctx, report); err != nil {
		return ReportView{}, err
	}
	return toReportView(report), nil
}

// GetByID returns a single report or ErrReportNotFound.
func (s *ReportService) GetByID(ctx context.Context, reportID string) (ReportView, error) {
	report, err := s.reports.FindByID(ctx, domain.ReportID(reportID))
	if err != nil {
		return ReportView{}, err
	}
	return toReportView(report), nil
}

// List returns reports for the authenticated inspector with optional pagination.
func (s *ReportService) List(ctx context.Context, inspectorID string, limit, offset int) ([]ReportView, error) {
	filter := domain.ReportFilter{Limit: limit, Offset: offset}
	reports, err := s.reports.FindByInspector(ctx, domain.InspectorID(inspectorID), filter)
	if err != nil {
		return nil, err
	}
	views := make([]ReportView, len(reports))
	for i, r := range reports {
		views[i] = toReportView(r)
	}
	return views, nil
}

// PDFPath returns the filesystem path of the generated PDF, or ErrReportNotGenerated.
func (s *ReportService) PDFPath(ctx context.Context, reportID string) (string, error) {
	report, err := s.reports.FindByID(ctx, domain.ReportID(reportID))
	if err != nil {
		return "", err
	}
	if report.PDFStoragePath == "" {
		return "", domain.ErrReportNotGenerated
	}
	return report.PDFStoragePath, nil
}
