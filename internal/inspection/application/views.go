package application

import "github.com/bejayjones/juno/internal/inspection/domain"

// InspectionView is the read model returned to REST clients.
type InspectionView struct {
	ID            string             `json:"id"`
	AppointmentID string             `json:"appointment_id"`
	InspectorID   string             `json:"inspector_id"`
	Status        string             `json:"status"`
	Header        HeaderView         `json:"header"`
	Systems       []SystemSectionView `json:"systems"`
	StartedAt     int64              `json:"started_at"`
	CompletedAt   *int64             `json:"completed_at"`
}

// HeaderView holds the required report header captured at walkthrough start.
type HeaderView struct {
	Weather       string   `json:"weather"`
	TemperatureF  int      `json:"temperature_f"`
	Attendees     []string `json:"attendees"`
	YearBuilt     int      `json:"year_built"`
	StructureType string   `json:"structure_type"`
}

// SystemSectionView is the read model for a single InterNACHI system.
type SystemSectionView struct {
	ID             string            `json:"id"`
	SystemType     string            `json:"system_type"`
	SystemLabel    string            `json:"system_label"`
	Descriptions   map[string]string `json:"descriptions"`
	Items          []ItemView        `json:"items"`
	InspectorNotes string            `json:"inspector_notes"`
	Progress       ProgressView      `json:"progress"`
	UpdatedAt      int64             `json:"updated_at"`
}

// ProgressView reports how many items have been addressed vs. total.
type ProgressView struct {
	Addressed int `json:"addressed"`
	Total     int `json:"total"`
}

// ItemView is the read model for one inspectable item.
type ItemView struct {
	ID                 string       `json:"id"`
	ItemKey            string       `json:"item_key"`
	Label              string       `json:"label"`
	Status             string       `json:"status"`
	NotInspectedReason string       `json:"not_inspected_reason,omitempty"`
	Findings           []FindingView `json:"findings"`
	UpdatedAt          int64        `json:"updated_at"`
}

// FindingView is the read model for a single finding.
type FindingView struct {
	ID           string         `json:"id"`
	Narrative    string         `json:"narrative"`
	IsDeficiency bool           `json:"is_deficiency"`
	Photos       []PhotoRefView `json:"photos"`
	CreatedAt    int64          `json:"created_at"`
	UpdatedAt    int64          `json:"updated_at"`
}

// PhotoRefView is the read model for a photo attached to a finding.
type PhotoRefView struct {
	ID          string `json:"id"`
	StoragePath string `json:"storage_path"`
	CapturedAt  int64  `json:"captured_at"`
}

// DeficiencyView is one entry in the aggregated deficiency summary.
type DeficiencyView struct {
	SystemType  string `json:"system_type"`
	SystemLabel string `json:"system_label"`
	ItemKey     string `json:"item_key"`
	ItemLabel   string `json:"item_label"`
	FindingID   string `json:"finding_id"`
	Narrative   string `json:"narrative"`
}

// ── Conversion helpers ────────────────────────────────────────────────────────

func toInspectionView(insp *domain.Inspection) InspectionView {
	v := InspectionView{
		ID:            string(insp.ID),
		AppointmentID: string(insp.AppointmentID),
		InspectorID:   string(insp.InspectorID),
		Status:        string(insp.Status),
		Header: HeaderView{
			Weather:       insp.Header.Weather,
			TemperatureF:  insp.Header.TemperatureF,
			Attendees:     insp.Header.Attendees,
			YearBuilt:     insp.Header.YearBuilt,
			StructureType: insp.Header.StructureType,
		},
		StartedAt: insp.StartedAt.Unix(),
	}
	if insp.CompletedAt != nil {
		t := insp.CompletedAt.Unix()
		v.CompletedAt = &t
	}

	// Emit systems in SOP order.
	for _, sysDef := range domain.Catalog {
		section, ok := insp.Systems[sysDef.Type]
		if !ok {
			continue
		}
		v.Systems = append(v.Systems, toSystemSectionView(section, sysDef))
	}
	return v
}

func toSystemSectionView(section *domain.SystemSection, sysDef *domain.SystemDefinition) SystemSectionView {
	addressed, total := section.Progress()

	descs := make(map[string]string, len(section.Descriptions))
	for k, val := range section.Descriptions {
		descs[string(k)] = val
	}

	items := make([]ItemView, len(section.Items))
	for i, item := range section.Items {
		items[i] = toItemView(item)
	}

	return SystemSectionView{
		ID:             section.ID,
		SystemType:     string(section.SystemType),
		SystemLabel:    sysDef.Label,
		Descriptions:   descs,
		Items:          items,
		InspectorNotes: section.InspectorNotes,
		Progress:       ProgressView{Addressed: addressed, Total: total},
		UpdatedAt:      section.UpdatedAt.Unix(),
	}
}

func toItemView(item domain.InspectionItem) ItemView {
	findings := make([]FindingView, len(item.Findings))
	for i, f := range item.Findings {
		findings[i] = toFindingView(f)
	}
	return ItemView{
		ID:                 item.ID,
		ItemKey:            string(item.ItemKey),
		Label:              item.Label,
		Status:             string(item.Status),
		NotInspectedReason: item.NotInspectedReason,
		Findings:           findings,
		UpdatedAt:          item.UpdatedAt.Unix(),
	}
}

func toFindingView(f domain.Finding) FindingView {
	photos := make([]PhotoRefView, len(f.Photos))
	for i, p := range f.Photos {
		photos[i] = PhotoRefView{
			ID:          string(p.ID),
			StoragePath: p.StoragePath,
			CapturedAt:  p.CapturedAt.Unix(),
		}
	}
	return FindingView{
		ID:           string(f.ID),
		Narrative:    f.Narrative,
		IsDeficiency: f.IsDeficiency,
		Photos:       photos,
		CreatedAt:    f.CreatedAt.Unix(),
		UpdatedAt:    f.UpdatedAt.Unix(),
	}
}

func toDeficiencyView(d domain.DeficiencySummaryItem) DeficiencyView {
	return DeficiencyView{
		SystemType:  string(d.SystemType),
		SystemLabel: d.SystemLabel,
		ItemKey:     string(d.ItemKey),
		ItemLabel:   d.ItemLabel,
		FindingID:   string(d.FindingID),
		Narrative:   d.Narrative,
	}
}
