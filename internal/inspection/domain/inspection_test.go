package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/bejayjones/juno/internal/inspection/domain"
)

var (
	testNow    = time.Date(2024, 1, 15, 9, 0, 0, 0, time.UTC)
	testHeader = domain.InspectionHeader{
		Weather:       "Clear",
		TemperatureF:  65,
		Attendees:     []string{"Client Name"},
		YearBuilt:     1995,
		StructureType: "Single family",
	}
)

var idSeq int

func nextID() string {
	idSeq++
	return "id-" + string(rune('0'+idSeq))
}

func newTestInspection(t *testing.T) *domain.Inspection {
	t.Helper()
	return domain.NewInspection("insp-1", "appt-1", "inspector-1", testHeader, nextID, testNow)
}

func TestNewInspection_InitializesAllTenSystems(t *testing.T) {
	insp := newTestInspection(t)

	if len(insp.Systems) != 10 {
		t.Fatalf("expected 10 systems, got %d", len(insp.Systems))
	}
	for _, st := range domain.AllSystems {
		if _, ok := insp.Systems[st]; !ok {
			t.Errorf("system %q not initialized", st)
		}
	}
}

func TestNewInspection_AllItemsStartUnaddressed(t *testing.T) {
	insp := newTestInspection(t)

	for _, section := range insp.Systems {
		for _, item := range section.Items {
			if item.IsAddressed() {
				t.Errorf("item %q should be unaddressed at start", item.ItemKey)
			}
		}
	}
}

func TestNewInspection_EmitsStartedEvent(t *testing.T) {
	insp := newTestInspection(t)

	events := insp.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	_, ok := events[0].(domain.InspectionStarted)
	if !ok {
		t.Errorf("expected InspectionStarted, got %T", events[0])
	}
}

func TestSetItemStatus_Inspected(t *testing.T) {
	insp := newTestInspection(t)

	err := insp.SetItemStatus(domain.SystemRoof, domain.ItemGuttersDownspouts, domain.StatusInspected, "", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	section := insp.Systems[domain.SystemRoof]
	item, _ := section.ItemByKey(domain.ItemGuttersDownspouts)
	if item.Status != domain.StatusInspected {
		t.Errorf("expected I, got %q", item.Status)
	}
}

func TestSetItemStatus_NI_RequiresReason(t *testing.T) {
	insp := newTestInspection(t)

	err := insp.SetItemStatus(domain.SystemRoof, domain.ItemGuttersDownspouts, domain.StatusNotInspected, "", testNow)
	if !errors.Is(err, domain.ErrNIReasonRequired) {
		t.Errorf("expected ErrNIReasonRequired, got %v", err)
	}
}

func TestSetItemStatus_NI_WithReason_Addressed(t *testing.T) {
	insp := newTestInspection(t)

	err := insp.SetItemStatus(domain.SystemRoof, domain.ItemGuttersDownspouts, domain.StatusNotInspected, "obscured by ice", testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	section := insp.Systems[domain.SystemRoof]
	item, _ := section.ItemByKey(domain.ItemGuttersDownspouts)
	if !item.IsAddressed() {
		t.Error("item with NI + reason should be addressed")
	}
}

func TestSetItemStatus_InvalidSystem(t *testing.T) {
	insp := newTestInspection(t)

	err := insp.SetItemStatus("nonexistent", domain.ItemGuttersDownspouts, domain.StatusInspected, "", testNow)
	if !errors.Is(err, domain.ErrInvalidSystemType) {
		t.Errorf("expected ErrInvalidSystemType, got %v", err)
	}
}

func TestSetItemStatus_InvalidItem(t *testing.T) {
	insp := newTestInspection(t)

	err := insp.SetItemStatus(domain.SystemRoof, "nonexistent.item", domain.StatusInspected, "", testNow)
	if !errors.Is(err, domain.ErrItemNotFound) {
		t.Errorf("expected ErrItemNotFound, got %v", err)
	}
}

func TestAddFinding_Deficiency_EmitsEvent(t *testing.T) {
	insp := newTestInspection(t)

	finding := domain.NewFinding("f-1", "Shingles missing at ridge", true, testNow)
	err := insp.AddFinding(domain.SystemRoof, domain.ItemGuttersDownspouts, finding, testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found bool
	for _, e := range insp.Events() {
		if _, ok := e.(domain.DeficiencyRecorded); ok {
			found = true
		}
	}
	if !found {
		t.Error("expected DeficiencyRecorded event")
	}
}

func TestComplete_FailsWhenItemsUnaddressed(t *testing.T) {
	insp := newTestInspection(t)

	err := insp.Complete(testNow)
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(ve.Fields) == 0 {
		t.Error("expected at least one missing field")
	}
}

func TestComplete_FailsWhenDescriptionsMissing(t *testing.T) {
	insp := newTestInspection(t)

	// Address every item so only descriptions are missing.
	for _, sysDef := range domain.Catalog {
		section := insp.Systems[sysDef.Type]
		for _, item := range section.Items {
			_ = insp.SetItemStatus(sysDef.Type, item.ItemKey, domain.StatusInspected, "", testNow)
		}
	}

	err := insp.Complete(testNow)
	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
}

func TestComplete_SucceedsWhenAllFieldsFilled(t *testing.T) {
	insp := fullyFilledInspection(t)

	err := insp.Complete(testNow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if insp.Status != domain.StatusCompleted {
		t.Errorf("expected completed, got %q", insp.Status)
	}
	if insp.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestComplete_CannotCompleteAlreadyCompleted(t *testing.T) {
	insp := fullyFilledInspection(t)
	_ = insp.Complete(testNow)

	err := insp.Complete(testNow.Add(time.Minute))
	if !errors.Is(err, domain.ErrInspectionCompleted) {
		t.Errorf("expected ErrInspectionCompleted, got %v", err)
	}
}

func TestDeficiencies_AggregatesAcrossSystems(t *testing.T) {
	insp := newTestInspection(t)

	_ = insp.AddFinding(domain.SystemRoof, domain.ItemGuttersDownspouts,
		domain.NewFinding("f-1", "Gutters clogged", true, testNow), testNow)
	_ = insp.AddFinding(domain.SystemElectrical, domain.ItemSmokeDetectors,
		domain.NewFinding("f-2", "No smoke detector in living room", true, testNow), testNow)
	_ = insp.AddFinding(domain.SystemElectrical, domain.ItemGFCIProtection,
		domain.NewFinding("f-3", "General observation (not a deficiency)", false, testNow), testNow)

	defs := insp.Deficiencies()
	if len(defs) != 2 {
		t.Errorf("expected 2 deficiencies, got %d", len(defs))
	}
}

// fullyFilledInspection sets every item to Inspected and fills all required
// descriptions, producing a valid inspection ready for Complete().
func fullyFilledInspection(t *testing.T) *domain.Inspection {
	t.Helper()
	insp := newTestInspection(t)

	for _, sysDef := range domain.Catalog {
		for _, item := range sysDef.Items {
			if err := insp.SetItemStatus(sysDef.Type, item.Key, domain.StatusInspected, "", testNow); err != nil {
				t.Fatalf("SetItemStatus: %v", err)
			}
		}
		for _, req := range sysDef.RequiredDescriptions {
			if err := insp.SetDescription(sysDef.Type, req.Key, "test value", testNow); err != nil {
				t.Fatalf("SetDescription: %v", err)
			}
		}
	}
	return insp
}
