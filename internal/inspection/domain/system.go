package domain

import "time"

// SystemType identifies one of the ten mandatory InterNACHI inspection systems.
type SystemType string

const (
	SystemRoof       SystemType = "roof"
	SystemExterior   SystemType = "exterior"
	SystemFoundation SystemType = "foundation"
	SystemHeating    SystemType = "heating"
	SystemCooling    SystemType = "cooling"
	SystemPlumbing   SystemType = "plumbing"
	SystemElectrical SystemType = "electrical"
	SystemFireplace  SystemType = "fireplace"
	SystemAttic      SystemType = "attic"
	SystemInterior   SystemType = "interior"
)

// AllSystems is the ordered list of the ten InterNACHI SOP inspection systems.
// Order matches SOP sections 3.1–3.10.
var AllSystems = []SystemType{
	SystemRoof,
	SystemExterior,
	SystemFoundation,
	SystemHeating,
	SystemCooling,
	SystemPlumbing,
	SystemElectrical,
	SystemFireplace,
	SystemAttic,
	SystemInterior,
}

// SystemSection is an entity within an Inspection representing one of the ten
// InterNACHI systems. It holds inspectable items and required description fields.
type SystemSection struct {
	ID             string
	SystemType     SystemType
	Items          []InspectionItem
	Descriptions   map[DescriptionKey]string // "shall describe" fields keyed by DescriptionKey
	InspectorNotes string
	UpdatedAt      time.Time
}

func (s *SystemSection) ItemByKey(key ItemKey) (*InspectionItem, error) {
	for i := range s.Items {
		if s.Items[i].ItemKey == key {
			return &s.Items[i], nil
		}
	}
	return nil, ErrItemNotFound
}

func (s *SystemSection) SetDescription(key DescriptionKey, value string, now time.Time) {
	if s.Descriptions == nil {
		s.Descriptions = make(map[DescriptionKey]string)
	}
	s.Descriptions[key] = value
	s.UpdatedAt = now
}

// Progress returns (addressed, total) counts for items in this section.
func (s *SystemSection) Progress() (addressed, total int) {
	total = len(s.Items)
	for _, item := range s.Items {
		if item.IsAddressed() {
			addressed++
		}
	}
	return
}

// MissingDescriptions returns required description keys that have not been filled.
func (s *SystemSection) MissingDescriptions(def *SystemDefinition) []DescriptionKey {
	var missing []DescriptionKey
	for _, req := range def.RequiredDescriptions {
		if v := s.Descriptions[req.Key]; v == "" {
			missing = append(missing, req.Key)
		}
	}
	return missing
}
