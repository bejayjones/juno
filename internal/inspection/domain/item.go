package domain

import "time"

// ItemStatus is the four-state InterNACHI condition rating for an inspection item.
type ItemStatus string

const (
	StatusInspected    ItemStatus = "I"
	StatusNotInspected ItemStatus = "NI"
	StatusNotPresent   ItemStatus = "NP"
	StatusDeficient    ItemStatus = "D"
)

// ItemKey is a canonical, stable identifier for an inspectable item in the catalog.
type ItemKey string

// DescriptionKey is the key for a required "shall describe" field on a SystemSection.
type DescriptionKey string

// InspectionItem is an entity within a SystemSection representing one inspectable
// component. Its status and findings are set by the inspector during the walkthrough.
type InspectionItem struct {
	ID                 string
	ItemKey            ItemKey
	Label              string
	Status             ItemStatus
	NotInspectedReason string
	Findings           []Finding
	UpdatedAt          time.Time
}

func newInspectionItem(id string, def ItemDefinition, now time.Time) InspectionItem {
	return InspectionItem{
		ID:        id,
		ItemKey:   def.Key,
		Label:     def.Label,
		Status:    StatusNotInspected,
		UpdatedAt: now,
	}
}

func (item *InspectionItem) SetStatus(status ItemStatus, reason string, now time.Time) error {
	if status == StatusNotInspected && reason == "" {
		return ErrNIReasonRequired
	}
	item.Status = status
	item.NotInspectedReason = reason
	item.UpdatedAt = now
	return nil
}

func (item *InspectionItem) AddFinding(f Finding) {
	item.Findings = append(item.Findings, f)
}

func (item *InspectionItem) UpdateFinding(id FindingID, narrative string, isDeficiency bool, now time.Time) error {
	for i := range item.Findings {
		if item.Findings[i].ID == id {
			item.Findings[i].Update(narrative, isDeficiency, now)
			return nil
		}
	}
	return ErrFindingNotFound
}

func (item *InspectionItem) RemoveFinding(id FindingID) error {
	for i, f := range item.Findings {
		if f.ID == id {
			item.Findings = append(item.Findings[:i], item.Findings[i+1:]...)
			return nil
		}
	}
	return ErrFindingNotFound
}

func (item *InspectionItem) FindingByID(id FindingID) (*Finding, error) {
	for i := range item.Findings {
		if item.Findings[i].ID == id {
			return &item.Findings[i], nil
		}
	}
	return nil, ErrFindingNotFound
}

// IsAddressed reports whether the inspector has explicitly set a status on this item.
// An item with StatusNotInspected and no reason is considered unaddressed (initial state).
func (item *InspectionItem) IsAddressed() bool {
	if item.Status == StatusNotInspected {
		return item.NotInspectedReason != ""
	}
	return true
}
