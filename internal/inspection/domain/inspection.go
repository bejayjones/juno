package domain

import (
	"fmt"
	"time"
)

type InspectionID string
type AppointmentID string
type InspectorID string

type InspectionStatus string

const (
	StatusInProgress InspectionStatus = "in_progress"
	StatusCompleted  InspectionStatus = "completed"
	StatusVoided     InspectionStatus = "voided"
)

// InspectionHeader holds the required report header fields captured at the
// start of a walkthrough (InterNACHI SOP required fields).
type InspectionHeader struct {
	Weather       string
	TemperatureF  int
	Attendees     []string
	YearBuilt     int
	StructureType string
}

// Inspection is the aggregate root for a walkthrough session. It owns all ten
// SystemSections and enforces InterNACHI completion rules.
type Inspection struct {
	ID            InspectionID
	AppointmentID AppointmentID
	InspectorID   InspectorID
	Status        InspectionStatus
	Header        InspectionHeader
	Systems       map[SystemType]*SystemSection
	StartedAt     time.Time
	CompletedAt   *time.Time
	events        []any
}

// NewInspection creates an Inspection initialized with all ten InterNACHI systems
// and their catalog items. idFn is called to generate IDs for the section and item
// entities, keeping the domain free of any UUID dependency.
func NewInspection(
	id InspectionID,
	appointmentID AppointmentID,
	inspectorID InspectorID,
	header InspectionHeader,
	idFn func() string,
	now time.Time,
) *Inspection {
	insp := &Inspection{
		ID:            id,
		AppointmentID: appointmentID,
		InspectorID:   inspectorID,
		Status:        StatusInProgress,
		Header:        header,
		Systems:       make(map[SystemType]*SystemSection),
		StartedAt:     now,
	}

	for _, sysDef := range Catalog {
		section := &SystemSection{
			ID:           idFn(),
			SystemType:   sysDef.Type,
			Descriptions: make(map[DescriptionKey]string),
			UpdatedAt:    now,
		}
		for _, itemDef := range sysDef.Items {
			section.Items = append(section.Items, newInspectionItem(idFn(), itemDef, now))
		}
		insp.Systems[sysDef.Type] = section
	}

	insp.record(InspectionStarted{
		InspectionID:  id,
		AppointmentID: appointmentID,
		InspectorID:   inspectorID,
		OccurredAt:    now,
	})

	return insp
}

func (i *Inspection) SetItemStatus(
	systemType SystemType, itemKey ItemKey,
	status ItemStatus, reason string,
	now time.Time,
) error {
	if i.Status != StatusInProgress {
		return ErrInspectionCompleted
	}
	section, item, err := i.locate(systemType, itemKey)
	if err != nil {
		return err
	}
	prev := item.Status
	if err := item.SetStatus(status, reason, now); err != nil {
		return err
	}
	section.UpdatedAt = now
	i.record(ItemStatusUpdated{
		InspectionID: i.ID,
		SystemType:   systemType,
		ItemKey:      itemKey,
		OldStatus:    prev,
		NewStatus:    status,
		OccurredAt:   now,
	})
	return nil
}

func (i *Inspection) SetDescription(
	systemType SystemType, key DescriptionKey, value string, now time.Time,
) error {
	if i.Status != StatusInProgress {
		return ErrInspectionCompleted
	}
	section, ok := i.Systems[systemType]
	if !ok {
		return ErrInvalidSystemType
	}
	section.SetDescription(key, value, now)
	return nil
}

func (i *Inspection) AddFinding(
	systemType SystemType, itemKey ItemKey, finding Finding, now time.Time,
) error {
	if i.Status != StatusInProgress {
		return ErrInspectionCompleted
	}
	section, item, err := i.locate(systemType, itemKey)
	if err != nil {
		return err
	}
	item.AddFinding(finding)
	section.UpdatedAt = now
	if finding.IsDeficiency {
		i.record(DeficiencyRecorded{
			InspectionID: i.ID,
			SystemType:   systemType,
			ItemKey:      itemKey,
			FindingID:    finding.ID,
			OccurredAt:   now,
		})
	}
	return nil
}

func (i *Inspection) UpdateFinding(
	systemType SystemType, itemKey ItemKey,
	findingID FindingID, narrative string, isDeficiency bool,
	now time.Time,
) error {
	if i.Status != StatusInProgress {
		return ErrInspectionCompleted
	}
	_, item, err := i.locate(systemType, itemKey)
	if err != nil {
		return err
	}
	return item.UpdateFinding(findingID, narrative, isDeficiency, now)
}

func (i *Inspection) RemoveFinding(
	systemType SystemType, itemKey ItemKey, findingID FindingID,
) error {
	if i.Status != StatusInProgress {
		return ErrInspectionCompleted
	}
	_, item, err := i.locate(systemType, itemKey)
	if err != nil {
		return err
	}
	return item.RemoveFinding(findingID)
}

// Complete validates all items are addressed and all required descriptions are
// filled, then marks the inspection as completed.
func (i *Inspection) Complete(now time.Time) error {
	if i.Status != StatusInProgress {
		return ErrInspectionCompleted
	}

	var missing []string
	for _, sysDef := range Catalog {
		section := i.Systems[sysDef.Type]
		for _, item := range section.Items {
			if !item.IsAddressed() {
				missing = append(missing,
					fmt.Sprintf("%s › %s: status not set", sysDef.Label, item.Label))
			}
		}
		for _, key := range section.MissingDescriptions(sysDef) {
			missing = append(missing,
				fmt.Sprintf("%s › %s: required description not filled", sysDef.Label, key))
		}
	}
	if len(missing) > 0 {
		return &ValidationError{Fields: missing}
	}

	t := now
	i.CompletedAt = &t
	i.Status = StatusCompleted
	i.record(InspectionCompleted{InspectionID: i.ID, OccurredAt: now})
	return nil
}

// DeficiencySummaryItem is a read model for one deficiency across the inspection.
type DeficiencySummaryItem struct {
	SystemType  SystemType
	SystemLabel string
	ItemKey     ItemKey
	ItemLabel   string
	FindingID   FindingID
	Narrative   string
}

// Deficiencies returns all deficient findings across all systems in SOP order.
func (i *Inspection) Deficiencies() []DeficiencySummaryItem {
	var out []DeficiencySummaryItem
	for _, sysDef := range Catalog {
		section := i.Systems[sysDef.Type]
		for _, item := range section.Items {
			for _, f := range item.Findings {
				if f.IsDeficiency {
					out = append(out, DeficiencySummaryItem{
						SystemType:  sysDef.Type,
						SystemLabel: sysDef.Label,
						ItemKey:     item.ItemKey,
						ItemLabel:   item.Label,
						FindingID:   f.ID,
						Narrative:   f.Narrative,
					})
				}
			}
		}
	}
	return out
}

func (i *Inspection) Events() []any    { return i.events }
func (i *Inspection) ClearEvents()     { i.events = nil }
func (i *Inspection) record(e any)     { i.events = append(i.events, e) }

// locate is a shared helper that looks up a section and item, returning typed errors.
func (i *Inspection) locate(systemType SystemType, itemKey ItemKey) (*SystemSection, *InspectionItem, error) {
	section, ok := i.Systems[systemType]
	if !ok {
		return nil, nil, ErrInvalidSystemType
	}
	item, err := section.ItemByKey(itemKey)
	if err != nil {
		return nil, nil, err
	}
	return section, item, nil
}
