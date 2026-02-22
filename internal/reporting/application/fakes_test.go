package application_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	inspectiondomain "github.com/bejayjones/juno/internal/inspection/domain"
	reportingapp "github.com/bejayjones/juno/internal/reporting/application"
	"github.com/bejayjones/juno/internal/reporting/domain"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/id"
)

// ── fakeReportRepo ────────────────────────────────────────────────────────────

type fakeReportRepo struct {
	byID           map[string]*domain.Report
	byInspectionID map[string]*domain.Report
}

func newFakeReportRepo() *fakeReportRepo {
	return &fakeReportRepo{
		byID:           make(map[string]*domain.Report),
		byInspectionID: make(map[string]*domain.Report),
	}
}

func (r *fakeReportRepo) Save(_ context.Context, report *domain.Report) error {
	r.byID[string(report.ID)] = report
	r.byInspectionID[string(report.InspectionID)] = report
	return nil
}

func (r *fakeReportRepo) FindByID(_ context.Context, rid domain.ReportID) (*domain.Report, error) {
	rep, ok := r.byID[string(rid)]
	if !ok {
		return nil, domain.ErrReportNotFound
	}
	return rep, nil
}

func (r *fakeReportRepo) FindByInspection(_ context.Context, inspectionID domain.InspectionID) (*domain.Report, error) {
	rep, ok := r.byInspectionID[string(inspectionID)]
	if !ok {
		return nil, domain.ErrReportNotFound
	}
	return rep, nil
}

func (r *fakeReportRepo) FindByInspector(_ context.Context, _ domain.InspectorID, _ domain.ReportFilter) ([]*domain.Report, error) {
	var out []*domain.Report
	for _, rep := range r.byID {
		out = append(out, rep)
	}
	return out, nil
}

func (r *fakeReportRepo) Delete(_ context.Context, rid domain.ReportID) error {
	rep, ok := r.byID[string(rid)]
	if !ok {
		return domain.ErrReportNotFound
	}
	delete(r.byInspectionID, string(rep.InspectionID))
	delete(r.byID, string(rid))
	return nil
}

// ── fakeInspectionRepo ────────────────────────────────────────────────────────

type fakeInspectionRepo struct {
	inspections map[string]*inspectiondomain.Inspection
}

func newFakeInspectionRepo() *fakeInspectionRepo {
	return &fakeInspectionRepo{inspections: make(map[string]*inspectiondomain.Inspection)}
}

func (r *fakeInspectionRepo) Save(_ context.Context, insp *inspectiondomain.Inspection) error {
	r.inspections[string(insp.ID)] = insp
	return nil
}

func (r *fakeInspectionRepo) FindByID(_ context.Context, inspID inspectiondomain.InspectionID) (*inspectiondomain.Inspection, error) {
	insp, ok := r.inspections[string(inspID)]
	if !ok {
		return nil, inspectiondomain.ErrInspectionNotFound
	}
	return insp, nil
}

func (r *fakeInspectionRepo) FindByAppointmentID(_ context.Context, apptID inspectiondomain.AppointmentID) (*inspectiondomain.Inspection, error) {
	for _, insp := range r.inspections {
		if insp.AppointmentID == apptID {
			return insp, nil
		}
	}
	return nil, inspectiondomain.ErrInspectionNotFound
}

func (r *fakeInspectionRepo) FindByInspector(_ context.Context, _ inspectiondomain.InspectorID, _ inspectiondomain.InspectionFilter) ([]*inspectiondomain.Inspection, error) {
	var out []*inspectiondomain.Inspection
	for _, insp := range r.inspections {
		out = append(out, insp)
	}
	return out, nil
}

func (r *fakeInspectionRepo) Delete(_ context.Context, inspID inspectiondomain.InspectionID) error {
	if _, ok := r.inspections[string(inspID)]; !ok {
		return inspectiondomain.ErrInspectionNotFound
	}
	delete(r.inspections, string(inspID))
	return nil
}

func (r *fakeInspectionRepo) FindPhotoMeta(_ context.Context, _ inspectiondomain.PhotoID) (string, string, error) {
	return "", "", inspectiondomain.ErrPhotoNotFound
}

// ── fakePDFGenerator ─────────────────────────────────────────────────────────

type fakePDFGenerator struct {
	err error
}

func (g *fakePDFGenerator) Generate(_ context.Context, _ *inspectiondomain.Inspection, outputPath string) error {
	if g.err != nil {
		return g.err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte("%PDF-1.4 fake"), 0600)
}

// ── fakeEmailSender ───────────────────────────────────────────────────────────

type fakeEmailSender struct {
	err error
}

func (s *fakeEmailSender) SendReport(_ context.Context, _, _ string) error {
	return s.err
}

// ── test helpers ──────────────────────────────────────────────────────────────

var fixedTime = time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)

func newSvc(t *testing.T, reportRepo *fakeReportRepo, inspRepo *fakeInspectionRepo, pdfErr error, emailErr error) *reportingapp.ReportService {
	t.Helper()
	return reportingapp.NewReportService(
		reportRepo,
		inspRepo,
		&fakePDFGenerator{err: pdfErr},
		&fakeEmailSender{err: emailErr},
		t.TempDir(),
		clock.Fixed(fixedTime),
	)
}

func seedInspection(t *testing.T, repo *fakeInspectionRepo) *inspectiondomain.Inspection {
	t.Helper()
	insp := inspectiondomain.NewInspection(
		inspectiondomain.InspectionID(id.New()),
		inspectiondomain.AppointmentID(id.New()),
		inspectiondomain.InspectorID(id.New()),
		inspectiondomain.InspectionHeader{
			Weather:       "Sunny",
			TemperatureF:  72,
			Attendees:     []string{"Alice"},
			YearBuilt:     1995,
			StructureType: "Single Family",
		},
		id.New,
		fixedTime,
	)
	_ = repo.Save(context.Background(), insp)
	return insp
}
