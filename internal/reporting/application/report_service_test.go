package application_test

import (
	"context"
	"errors"
	"testing"

	reportingapp "github.com/bejayjones/juno/internal/reporting/application"
	"github.com/bejayjones/juno/internal/reporting/domain"
	"github.com/bejayjones/juno/pkg/clock"
)

func TestGenerateReport_HappyPath(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	svc := newSvc(t, reportRepo, inspRepo, nil, reportingapp.ErrDeliveryQueued)

	view, err := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}
	if view.ID == "" {
		t.Error("expected non-empty report ID")
	}
	if view.Status != string(domain.ReportDraft) {
		t.Errorf("status = %q, want %q", view.Status, domain.ReportDraft)
	}
	if view.GeneratedAt == nil {
		t.Error("expected GeneratedAt to be set")
	}
	if view.PDFStoragePath == "" {
		t.Error("expected PDFStoragePath to be set")
	}
}

func TestGenerateReport_AlreadyExists(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	svc := newSvc(t, reportRepo, inspRepo, nil, reportingapp.ErrDeliveryQueued)

	_, err := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if err != nil {
		t.Fatalf("first GenerateReport: %v", err)
	}

	_, err = svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if !errors.Is(err, domain.ErrReportAlreadyExists) {
		t.Errorf("want ErrReportAlreadyExists, got %v", err)
	}
}

func TestGenerateReport_PDFError(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	pdfErr := errors.New("fpdf failure")
	svc := newSvc(t, reportRepo, inspRepo, pdfErr, reportingapp.ErrDeliveryQueued)

	_, err := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if err == nil {
		t.Fatal("expected error from PDF generator, got nil")
	}
}

func TestFinalizeReport(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	svc := newSvc(t, reportRepo, inspRepo, nil, reportingapp.ErrDeliveryQueued)

	view, err := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	view, err = svc.FinalizeReport(context.Background(), view.ID)
	if err != nil {
		t.Fatalf("FinalizeReport: %v", err)
	}
	if view.Status != string(domain.ReportFinalized) {
		t.Errorf("status = %q, want %q", view.Status, domain.ReportFinalized)
	}
}

func TestFinalizeReport_NotGenerated(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	// Manually save a draft report with no PDF path.
	svc := newSvc(t, reportRepo, inspRepo, nil, reportingapp.ErrDeliveryQueued)
	view, _ := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))

	// Simulate a report that never had PDF generated (manually wipe pdf path).
	rep, _ := reportRepo.FindByID(context.Background(), domain.ReportID(view.ID))
	rep.PDFStoragePath = ""
	rep.GeneratedAt = nil

	_, err := svc.FinalizeReport(context.Background(), view.ID)
	if !errors.Is(err, domain.ErrReportNotGenerated) {
		t.Errorf("want ErrReportNotGenerated, got %v", err)
	}
}

func TestQueueDelivery_QueueOnly(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	// Email sender returns ErrDeliveryQueued (queue-only mode).
	svc := newSvc(t, reportRepo, inspRepo, nil, reportingapp.ErrDeliveryQueued)

	reportView, err := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	dv, err := svc.QueueDelivery(context.Background(), reportView.ID, "client@example.com")
	if err != nil {
		t.Fatalf("QueueDelivery: %v", err)
	}
	if dv.Status != string(domain.DeliveryPending) {
		t.Errorf("delivery status = %q, want %q", dv.Status, domain.DeliveryPending)
	}
}

func TestQueueDelivery_SMTPSuccess(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	// Email sender returns nil = success.
	svc := newSvc(t, reportRepo, inspRepo, nil, nil)

	reportView, err := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	dv, err := svc.QueueDelivery(context.Background(), reportView.ID, "client@example.com")
	if err != nil {
		t.Fatalf("QueueDelivery: %v", err)
	}
	if dv.Status != string(domain.DeliverySent) {
		t.Errorf("delivery status = %q, want %q", dv.Status, domain.DeliverySent)
	}
	if dv.SentAt == nil {
		t.Error("expected SentAt to be set after successful send")
	}
}

func TestQueueDelivery_SMTPFailure(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	svc := newSvc(t, reportRepo, inspRepo, nil, errors.New("connection refused"))

	reportView, err := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	if err != nil {
		t.Fatalf("GenerateReport: %v", err)
	}

	dv, err := svc.QueueDelivery(context.Background(), reportView.ID, "client@example.com")
	if err != nil {
		t.Fatalf("QueueDelivery: %v", err)
	}
	if dv.Status != string(domain.DeliveryFailed_) {
		t.Errorf("delivery status = %q, want %q", dv.Status, domain.DeliveryFailed_)
	}
	if dv.FailureReason == "" {
		t.Error("expected non-empty failure reason")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := newSvc(t, newFakeReportRepo(), newFakeInspectionRepo(), nil, reportingapp.ErrDeliveryQueued)

	_, err := svc.GetByID(context.Background(), "non-existent")
	if !errors.Is(err, domain.ErrReportNotFound) {
		t.Errorf("want ErrReportNotFound, got %v", err)
	}
}

func TestList(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()

	inspA := seedInspection(t, inspRepo)
	inspB := seedInspection(t, inspRepo)

	svc := newSvc(t, reportRepo, inspRepo, nil, reportingapp.ErrDeliveryQueued)

	_, _ = svc.GenerateReport(context.Background(), string(inspA.ID), string(inspA.InspectorID))
	_, _ = svc.GenerateReport(context.Background(), string(inspB.ID), string(inspB.InspectorID))

	views, err := svc.List(context.Background(), string(inspA.InspectorID), 50, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(views) == 0 {
		t.Error("expected at least one report in list")
	}
}

func TestRetryFailedDeliveries(t *testing.T) {
	reportRepo := newFakeReportRepo()
	inspRepo := newFakeInspectionRepo()
	insp := seedInspection(t, inspRepo)

	// First call: email fails.
	svc := newSvc(t, reportRepo, inspRepo, nil, errors.New("smtp down"))
	reportView, _ := svc.GenerateReport(context.Background(), string(insp.ID), string(insp.InspectorID))
	_, _ = svc.QueueDelivery(context.Background(), reportView.ID, "client@example.com")

	// Retry with a new service that has a working email sender.
	svc2 := reportingapp.NewReportService(
		reportRepo,
		inspRepo,
		&fakePDFGenerator{},
		&fakeEmailSender{err: nil}, // now succeeds
		t.TempDir(),
		clock.Fixed(fixedTime),
	)

	view, err := svc2.RetryFailedDeliveries(context.Background(), reportView.ID)
	if err != nil {
		t.Fatalf("RetryFailedDeliveries: %v", err)
	}
	if len(view.Deliveries) == 0 {
		t.Fatal("expected deliveries in report view")
	}
	if view.Deliveries[0].Status != string(domain.DeliverySent) {
		t.Errorf("delivery status after retry = %q, want %q", view.Deliveries[0].Status, domain.DeliverySent)
	}
}
