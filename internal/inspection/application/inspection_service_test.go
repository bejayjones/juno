package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/bejayjones/juno/internal/inspection/application"
	"github.com/bejayjones/juno/internal/inspection/domain"
	"github.com/bejayjones/juno/pkg/clock"
)

func newSvc(t *testing.T) *application.InspectionService {
	t.Helper()
	return application.NewInspectionService(newFakeRepo(), clock.Fixed(time.Date(2025, 6, 1, 10, 0, 0, 0, time.UTC)))
}

func startTestInspection(t *testing.T, svc *application.InspectionService) application.InspectionView {
	t.Helper()
	view, err := svc.Start(context.Background(), application.StartInput{
		AppointmentID: "appt-1",
		InspectorID:   "insp-1",
		Weather:       "Clear",
		TemperatureF:  72,
		Attendees:     []string{"Jane Doe"},
		YearBuilt:     1995,
		StructureType: "Single Family",
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	return view
}

func TestStart_InitializesAllTenSystems(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	if len(view.Systems) != 10 {
		t.Fatalf("want 10 systems, got %d", len(view.Systems))
	}
	if view.Status != "in_progress" {
		t.Errorf("want status in_progress, got %s", view.Status)
	}
	if view.AppointmentID != "appt-1" {
		t.Errorf("want appointment_id appt-1, got %s", view.AppointmentID)
	}
}

func TestStart_AllItemsUnaddressed(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	for _, sys := range view.Systems {
		if sys.Progress.Addressed != 0 {
			t.Errorf("system %s: want 0 addressed items, got %d", sys.SystemType, sys.Progress.Addressed)
		}
		if sys.Progress.Total == 0 {
			t.Errorf("system %s: want >0 total items", sys.SystemType)
		}
	}
}

func TestSetItemStatus_Success(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	updated, err := svc.SetItemStatus(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts", "I", "")
	if err != nil {
		t.Fatalf("SetItemStatus: %v", err)
	}

	var roofSys *application.SystemSectionView
	for i := range updated.Systems {
		if updated.Systems[i].SystemType == "roof" {
			roofSys = &updated.Systems[i]
			break
		}
	}
	if roofSys == nil {
		t.Fatal("roof system not found")
	}

	var found bool
	for _, item := range roofSys.Items {
		if item.ItemKey == "roof.gutters_downspouts" {
			if item.Status != "I" {
				t.Errorf("want status I, got %s", item.Status)
			}
			found = true
		}
	}
	if !found {
		t.Error("item roof.gutters_downspouts not found in roof system")
	}
}

func TestSetItemStatus_NI_RequiresReason(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	_, err := svc.SetItemStatus(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts", "NI", "")
	if err == nil {
		t.Fatal("expected error for NI without reason, got nil")
	}
}

func TestSetItemStatus_NI_WithReason(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	_, err := svc.SetItemStatus(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts", "NI", "snow covered")
	if err != nil {
		t.Fatalf("SetItemStatus NI with reason: %v", err)
	}
}

func TestAddFinding_Success(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	f, err := svc.AddFinding(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts",
		application.AddFindingInput{
			Narrative:    "Missing gutter section on north side",
			IsDeficiency: true,
		})
	if err != nil {
		t.Fatalf("AddFinding: %v", err)
	}
	if f.Narrative != "Missing gutter section on north side" {
		t.Errorf("unexpected narrative: %s", f.Narrative)
	}
	if !f.IsDeficiency {
		t.Error("want IsDeficiency=true")
	}
}

func TestUpdateFinding(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	f, _ := svc.AddFinding(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts",
		application.AddFindingInput{Narrative: "old", IsDeficiency: false})

	updated, err := svc.UpdateFinding(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts", f.ID,
		application.UpdateFindingInput{Narrative: "new narrative", IsDeficiency: true})
	if err != nil {
		t.Fatalf("UpdateFinding: %v", err)
	}
	if updated.Narrative != "new narrative" {
		t.Errorf("want narrative 'new narrative', got %s", updated.Narrative)
	}
}

func TestDeleteFinding(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	f, _ := svc.AddFinding(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts",
		application.AddFindingInput{Narrative: "to delete", IsDeficiency: false})

	if err := svc.DeleteFinding(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts", f.ID); err != nil {
		t.Fatalf("DeleteFinding: %v", err)
	}

	// Verify finding is gone.
	got, _ := svc.GetByID(context.Background(), view.ID)
	for _, sys := range got.Systems {
		if sys.SystemType == "roof" {
			for _, item := range sys.Items {
				if item.ItemKey == "roof.gutters_downspouts" {
					for _, finding := range item.Findings {
						if finding.ID == f.ID {
							t.Error("finding still present after delete")
						}
					}
				}
			}
		}
	}
}

func TestComplete_FailsWhenItemsUnaddressed(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	_, err := svc.Complete(context.Background(), view.ID)
	if err == nil {
		t.Fatal("expected ValidationError, got nil")
	}
	var ve *domain.ValidationError
	if !isValidationError(err, &ve) {
		t.Fatalf("expected ValidationError, got %T: %v", err, err)
	}
	if len(ve.Fields) == 0 {
		t.Error("expected non-empty Fields in ValidationError")
	}
}

func TestGetDeficiencySummary(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	svc.AddFinding(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts",
		application.AddFindingInput{Narrative: "deficiency 1", IsDeficiency: true})

	svc.AddFinding(context.Background(),
		view.ID, "roof", "roof.gutters_downspouts",
		application.AddFindingInput{Narrative: "not a deficiency", IsDeficiency: false})

	defs, err := svc.GetDeficiencySummary(context.Background(), view.ID)
	if err != nil {
		t.Fatalf("GetDeficiencySummary: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("want 1 deficiency, got %d", len(defs))
	}
	if defs[0].Narrative != "deficiency 1" {
		t.Errorf("unexpected narrative: %s", defs[0].Narrative)
	}
}

func TestGetSystemSection(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	sys, err := svc.GetSystemSection(context.Background(), view.ID, "electrical")
	if err != nil {
		t.Fatalf("GetSystemSection: %v", err)
	}
	if sys.SystemType != "electrical" {
		t.Errorf("want electrical, got %s", sys.SystemType)
	}
	if len(sys.Items) == 0 {
		t.Error("electrical system has no items")
	}
}

func TestSetDescriptions(t *testing.T) {
	svc := newSvc(t)
	view := startTestInspection(t, svc)

	sys, err := svc.SetDescriptions(context.Background(), view.ID, "roof", map[string]string{
		"roof.covering_material": "Asphalt shingles",
	})
	if err != nil {
		t.Fatalf("SetDescriptions: %v", err)
	}
	if sys.Descriptions["roof.covering_material"] != "Asphalt shingles" {
		t.Errorf("description not set correctly: %v", sys.Descriptions)
	}
}

func TestGetByAppointmentID(t *testing.T) {
	svc := newSvc(t)
	startTestInspection(t, svc)

	view, err := svc.GetByAppointmentID(context.Background(), "appt-1")
	if err != nil {
		t.Fatalf("GetByAppointmentID: %v", err)
	}
	if view.AppointmentID != "appt-1" {
		t.Errorf("want appt-1, got %s", view.AppointmentID)
	}
}

func TestList(t *testing.T) {
	svc := newSvc(t)
	startTestInspection(t, svc)

	// Also start a second inspection for a different inspector.
	svc.Start(context.Background(), application.StartInput{
		AppointmentID: "appt-2",
		InspectorID:   "insp-2",
		Weather:       "Rainy",
	})

	views, err := svc.List(context.Background(), "insp-1", domain.InspectionFilter{Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(views) != 1 {
		t.Errorf("want 1 inspection for insp-1, got %d", len(views))
	}
}

// isValidationError checks if err is a *domain.ValidationError and populates ve.
func isValidationError(err error, ve **domain.ValidationError) bool {
	if e, ok := err.(*domain.ValidationError); ok {
		*ve = e
		return true
	}
	return false
}
