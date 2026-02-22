package application

import (
	"context"
	"fmt"
	"io"

	"github.com/bejayjones/juno/internal/inspection/domain"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/id"
	"github.com/bejayjones/juno/pkg/storage"
)

// InspectionService handles all walkthrough operations.
type InspectionService struct {
	inspections domain.InspectionRepository
	photos      storage.PhotoStorage
	clock       clock.Clock
}

func NewInspectionService(
	inspections domain.InspectionRepository,
	photos storage.PhotoStorage,
	clk clock.Clock,
) *InspectionService {
	return &InspectionService{inspections: inspections, photos: photos, clock: clk}
}

// ── Commands ──────────────────────────────────────────────────────────────────

// StartInput holds the required fields to begin an inspection.
type StartInput struct {
	AppointmentID string
	InspectorID   string
	Weather       string
	TemperatureF  int
	Attendees     []string
	YearBuilt     int
	StructureType string
}

// Start creates a new inspection initialized with all ten InterNACHI systems.
func (s *InspectionService) Start(ctx context.Context, in StartInput) (InspectionView, error) {
	now := s.clock.Now()
	insp := domain.NewInspection(
		domain.InspectionID(id.New()),
		domain.AppointmentID(in.AppointmentID),
		domain.InspectorID(in.InspectorID),
		domain.InspectionHeader{
			Weather:       in.Weather,
			TemperatureF:  in.TemperatureF,
			Attendees:     in.Attendees,
			YearBuilt:     in.YearBuilt,
			StructureType: in.StructureType,
		},
		id.New, // ID generator for section and item entities
		now,
	)
	if err := s.inspections.Save(ctx, insp); err != nil {
		return InspectionView{}, fmt.Errorf("save inspection: %w", err)
	}
	return toInspectionView(insp), nil
}

// SetItemStatus updates an item's I/NI/NP/D status.
func (s *InspectionService) SetItemStatus(
	ctx context.Context,
	inspID, systemType, itemKey, status, reason string,
) (InspectionView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return InspectionView{}, err
	}
	if err := insp.SetItemStatus(
		domain.SystemType(systemType),
		domain.ItemKey(itemKey),
		domain.ItemStatus(status),
		reason,
		s.clock.Now(),
	); err != nil {
		return InspectionView{}, err
	}
	if err := s.inspections.Save(ctx, insp); err != nil {
		return InspectionView{}, fmt.Errorf("save inspection: %w", err)
	}
	return toInspectionView(insp), nil
}

// SetDescriptions saves one or more "shall describe" fields for a system.
func (s *InspectionService) SetDescriptions(
	ctx context.Context,
	inspID, systemType string,
	descriptions map[string]string,
) (SystemSectionView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return SystemSectionView{}, err
	}
	now := s.clock.Now()
	for key, value := range descriptions {
		if err := insp.SetDescription(
			domain.SystemType(systemType),
			domain.DescriptionKey(key),
			value,
			now,
		); err != nil {
			return SystemSectionView{}, err
		}
	}
	if err := s.inspections.Save(ctx, insp); err != nil {
		return SystemSectionView{}, fmt.Errorf("save inspection: %w", err)
	}
	section := insp.Systems[domain.SystemType(systemType)]
	sysDef := domain.CatalogBySystem(domain.SystemType(systemType))
	return toSystemSectionView(section, sysDef), nil
}

// AddFindingInput holds the fields for a new finding.
type AddFindingInput struct {
	Narrative    string
	IsDeficiency bool
}

// AddFinding creates a new finding on an item.
func (s *InspectionService) AddFinding(
	ctx context.Context,
	inspID, systemType, itemKey string,
	in AddFindingInput,
) (FindingView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return FindingView{}, err
	}
	now := s.clock.Now()
	finding := domain.NewFinding(
		domain.FindingID(id.New()),
		in.Narrative,
		in.IsDeficiency,
		now,
	)
	if err := insp.AddFinding(
		domain.SystemType(systemType),
		domain.ItemKey(itemKey),
		finding,
		now,
	); err != nil {
		return FindingView{}, err
	}
	if err := s.inspections.Save(ctx, insp); err != nil {
		return FindingView{}, fmt.Errorf("save inspection: %w", err)
	}
	return toFindingView(finding), nil
}

// UpdateFindingInput holds the mutable fields of a finding.
type UpdateFindingInput struct {
	Narrative    string
	IsDeficiency bool
}

// UpdateFinding modifies narrative and deficiency flag of an existing finding.
func (s *InspectionService) UpdateFinding(
	ctx context.Context,
	inspID, systemType, itemKey, findingID string,
	in UpdateFindingInput,
) (FindingView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return FindingView{}, err
	}
	now := s.clock.Now()
	if err := insp.UpdateFinding(
		domain.SystemType(systemType),
		domain.ItemKey(itemKey),
		domain.FindingID(findingID),
		in.Narrative,
		in.IsDeficiency,
		now,
	); err != nil {
		return FindingView{}, err
	}
	if err := s.inspections.Save(ctx, insp); err != nil {
		return FindingView{}, fmt.Errorf("save inspection: %w", err)
	}

	// Retrieve updated finding view from the saved aggregate.
	section := insp.Systems[domain.SystemType(systemType)]
	item, _ := section.ItemByKey(domain.ItemKey(itemKey))
	f, _ := item.FindingByID(domain.FindingID(findingID))
	return toFindingView(*f), nil
}

// DeleteFinding removes a finding from an item.
func (s *InspectionService) DeleteFinding(
	ctx context.Context,
	inspID, systemType, itemKey, findingID string,
) error {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return err
	}
	if err := insp.RemoveFinding(
		domain.SystemType(systemType),
		domain.ItemKey(itemKey),
		domain.FindingID(findingID),
	); err != nil {
		return err
	}
	if err := s.inspections.Save(ctx, insp); err != nil {
		return fmt.Errorf("save inspection: %w", err)
	}
	return nil
}

// Complete validates and finalizes the inspection.
func (s *InspectionService) Complete(ctx context.Context, inspID string) (InspectionView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return InspectionView{}, err
	}
	if err := insp.Complete(s.clock.Now()); err != nil {
		return InspectionView{}, err
	}
	if err := s.inspections.Save(ctx, insp); err != nil {
		return InspectionView{}, fmt.Errorf("save inspection: %w", err)
	}
	return toInspectionView(insp), nil
}

// ── Queries ───────────────────────────────────────────────────────────────────

// GetByID loads a single inspection.
func (s *InspectionService) GetByID(ctx context.Context, inspID string) (InspectionView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return InspectionView{}, err
	}
	return toInspectionView(insp), nil
}

// GetByAppointmentID loads the inspection linked to an appointment.
func (s *InspectionService) GetByAppointmentID(ctx context.Context, appointmentID string) (InspectionView, error) {
	insp, err := s.inspections.FindByAppointmentID(ctx, domain.AppointmentID(appointmentID))
	if err != nil {
		return InspectionView{}, err
	}
	return toInspectionView(insp), nil
}

// List returns inspections for the given inspector.
func (s *InspectionService) List(ctx context.Context, inspectorID string, filter domain.InspectionFilter) ([]InspectionView, error) {
	inspections, err := s.inspections.FindByInspector(ctx, domain.InspectorID(inspectorID), filter)
	if err != nil {
		return nil, fmt.Errorf("list inspections: %w", err)
	}
	views := make([]InspectionView, len(inspections))
	for i, insp := range inspections {
		views[i] = toInspectionView(insp)
	}
	return views, nil
}

// GetSystemSection returns the view for a single system within an inspection.
func (s *InspectionService) GetSystemSection(
	ctx context.Context,
	inspID, systemType string,
) (SystemSectionView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return SystemSectionView{}, err
	}
	section, ok := insp.Systems[domain.SystemType(systemType)]
	if !ok {
		return SystemSectionView{}, domain.ErrInvalidSystemType
	}
	sysDef := domain.CatalogBySystem(domain.SystemType(systemType))
	return toSystemSectionView(section, sysDef), nil
}

// GetDeficiencySummary returns all deficient findings across the inspection.
func (s *InspectionService) GetDeficiencySummary(ctx context.Context, inspID string) ([]DeficiencyView, error) {
	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return nil, err
	}
	defs := insp.Deficiencies()
	views := make([]DeficiencyView, len(defs))
	for i, d := range defs {
		views[i] = toDeficiencyView(d)
	}
	return views, nil
}

// ── Photo operations ──────────────────────────────────────────────────────────

// AddPhoto saves a photo file to storage and attaches it to the given finding.
// mimeType must be one of the values in storage.AllowedMimeTypes.
func (s *InspectionService) AddPhoto(
	ctx context.Context,
	inspID, systemType, itemKey, findingID string,
	mimeType string,
	data io.Reader,
) (PhotoRefView, error) {
	ext, ok := storage.AllowedMimeTypes[mimeType]
	if !ok {
		return PhotoRefView{}, domain.ErrInvalidMimeType
	}

	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return PhotoRefView{}, err
	}

	section, ok := insp.Systems[domain.SystemType(systemType)]
	if !ok {
		return PhotoRefView{}, domain.ErrInvalidSystemType
	}
	item, err := section.ItemByKey(domain.ItemKey(itemKey))
	if err != nil {
		return PhotoRefView{}, err
	}
	finding, err := item.FindingByID(domain.FindingID(findingID))
	if err != nil {
		return PhotoRefView{}, err
	}

	photoID := id.New()
	storagePath, err := s.photos.Save(ctx, photoID, ext, data)
	if err != nil {
		return PhotoRefView{}, fmt.Errorf("save photo to storage: %w", err)
	}

	now := s.clock.Now()
	ref := domain.PhotoRef{
		ID:          domain.PhotoID(photoID),
		StoragePath: storagePath,
		MimeType:    mimeType,
		CapturedAt:  now,
	}
	finding.AddPhoto(ref)

	if err := s.inspections.Save(ctx, insp); err != nil {
		// Best-effort cleanup: remove the orphaned file.
		_ = s.photos.Delete(ctx, storagePath)
		return PhotoRefView{}, fmt.Errorf("save inspection: %w", err)
	}

	return PhotoRefView{
		ID:          photoID,
		StoragePath: storagePath,
		MimeType:    mimeType,
		CapturedAt:  now.Unix(),
	}, nil
}

// DeletePhoto removes a photo from its finding and from storage.
func (s *InspectionService) DeletePhoto(
	ctx context.Context,
	inspID, systemType, itemKey, photoID string,
) error {
	// Fetch storage path first — it's available in the DB before we save.
	storagePath, _, err := s.inspections.FindPhotoMeta(ctx, domain.PhotoID(photoID))
	if err != nil {
		return err // ErrPhotoNotFound or a DB error
	}

	insp, err := s.inspections.FindByID(ctx, domain.InspectionID(inspID))
	if err != nil {
		return err
	}

	section, ok := insp.Systems[domain.SystemType(systemType)]
	if !ok {
		return domain.ErrInvalidSystemType
	}
	item, err := section.ItemByKey(domain.ItemKey(itemKey))
	if err != nil {
		return err
	}

	// Remove from the in-memory aggregate (search all findings on the item).
	var found bool
	for i := range item.Findings {
		if item.Findings[i].RemovePhoto(domain.PhotoID(photoID)) {
			found = true
			break
		}
	}
	if !found {
		return domain.ErrPhotoNotFound
	}

	// Persist — findings delete+reinsert removes the photo row from the DB.
	if err := s.inspections.Save(ctx, insp); err != nil {
		return fmt.Errorf("save inspection: %w", err)
	}

	// Best-effort storage cleanup (don't fail the request on storage error).
	_ = s.photos.Delete(ctx, storagePath)
	return nil
}

// GetPhotoData returns a ReadCloser and MIME type for streaming a photo to a client.
func (s *InspectionService) GetPhotoData(ctx context.Context, photoID string) (io.ReadCloser, string, error) {
	storagePath, mimeType, err := s.inspections.FindPhotoMeta(ctx, domain.PhotoID(photoID))
	if err != nil {
		return nil, "", err
	}
	rc, err := s.photos.Get(ctx, storagePath)
	if err != nil {
		return nil, "", fmt.Errorf("get photo from storage: %w", err)
	}
	return rc, mimeType, nil
}
