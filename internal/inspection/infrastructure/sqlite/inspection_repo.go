// Package sqlite provides the SQLite-backed repository for the inspection context.
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bejayjones/juno/internal/inspection/domain"
	"github.com/bejayjones/juno/internal/platform/db"
	"github.com/bejayjones/juno/internal/sync/recorder"
)

// InspectionRepository persists Inspection aggregates in SQLite.
type InspectionRepository struct {
	db       *db.DB
	recorder *recorder.Recorder
}

func NewInspectionRepository(database *db.DB) *InspectionRepository {
	return &InspectionRepository{db: database}
}

// WithRecorder enables sync recording for this repository.
func (r *InspectionRepository) WithRecorder(rec *recorder.Recorder) *InspectionRepository {
	r.recorder = rec
	return r
}

// Save upserts the entire Inspection aggregate tree within a single transaction.
func (r *InspectionRepository) Save(ctx context.Context, insp *domain.Inspection) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if err := saveInspection(tx, insp); err != nil {
			return err
		}
		return r.recorder.Record(ctx, tx, "inspections", string(insp.ID), "upsert", insp)
	})
}

func saveInspection(tx *sql.Tx, insp *domain.Inspection) error {
	attendeesJSON, err := json.Marshal(insp.Header.Attendees)
	if err != nil {
		return fmt.Errorf("marshal attendees: %w", err)
	}

	var completedAt *int64
	if insp.CompletedAt != nil {
		t := insp.CompletedAt.Unix()
		completedAt = &t
	}

	_, err = tx.Exec(`
		INSERT INTO inspections
			(id, appointment_id, inspector_id, status,
			 weather, temperature_f, attendees, year_built, structure_type,
			 started_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status         = excluded.status,
			weather        = excluded.weather,
			temperature_f  = excluded.temperature_f,
			attendees      = excluded.attendees,
			year_built     = excluded.year_built,
			structure_type = excluded.structure_type,
			completed_at   = excluded.completed_at,
			updated_at     = excluded.updated_at
	`,
		string(insp.ID), string(insp.AppointmentID), string(insp.InspectorID),
		string(insp.Status),
		insp.Header.Weather, insp.Header.TemperatureF,
		string(attendeesJSON),
		insp.Header.YearBuilt, insp.Header.StructureType,
		insp.StartedAt.Unix(), completedAt,
		insp.StartedAt.Unix(), // created_at: set once on creation, not updated
		time.Now().Unix(),     // updated_at
	)
	if err != nil {
		return fmt.Errorf("upsert inspection: %w", err)
	}

	for _, section := range insp.Systems {
		if err := saveSystemSection(tx, string(insp.ID), section); err != nil {
			return err
		}
	}
	return nil
}

func saveSystemSection(tx *sql.Tx, inspectionID string, section *domain.SystemSection) error {
	_, err := tx.Exec(`
		INSERT INTO system_sections (id, inspection_id, system_type, inspector_notes, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			inspector_notes = excluded.inspector_notes,
			updated_at      = excluded.updated_at
	`,
		section.ID, inspectionID, string(section.SystemType),
		section.InspectorNotes, section.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("upsert system_section %s: %w", section.SystemType, err)
	}

	// Delete then re-insert descriptions (map may have grown or values changed).
	if _, err := tx.Exec(`DELETE FROM system_descriptions WHERE system_section_id = ?`, section.ID); err != nil {
		return fmt.Errorf("delete descriptions %s: %w", section.SystemType, err)
	}
	for key, value := range section.Descriptions {
		if _, err := tx.Exec(
			`INSERT INTO system_descriptions (system_section_id, description_key, value) VALUES (?, ?, ?)`,
			section.ID, string(key), value,
		); err != nil {
			return fmt.Errorf("insert description %s.%s: %w", section.SystemType, key, err)
		}
	}

	for i := range section.Items {
		if err := saveInspectionItem(tx, section.ID, &section.Items[i]); err != nil {
			return err
		}
	}
	return nil
}

func saveInspectionItem(tx *sql.Tx, sectionID string, item *domain.InspectionItem) error {
	_, err := tx.Exec(`
		INSERT INTO inspection_items
			(id, system_section_id, item_key, label, status, not_inspected_reason, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status               = excluded.status,
			not_inspected_reason = excluded.not_inspected_reason,
			updated_at           = excluded.updated_at
	`,
		item.ID, sectionID, string(item.ItemKey), item.Label,
		string(item.Status), item.NotInspectedReason, item.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("upsert item %s: %w", item.ItemKey, err)
	}

	// Delete then re-insert findings (handles both adds and removes).
	// Cascade delete also removes photos for deleted findings.
	if _, err := tx.Exec(`DELETE FROM findings WHERE inspection_item_id = ?`, item.ID); err != nil {
		return fmt.Errorf("delete findings for item %s: %w", item.ItemKey, err)
	}
	for _, f := range item.Findings {
		if err := saveFinding(tx, item.ID, &f); err != nil {
			return err
		}
	}
	return nil
}

func saveFinding(tx *sql.Tx, itemID string, f *domain.Finding) error {
	_, err := tx.Exec(`
		INSERT INTO findings (id, inspection_item_id, narrative, is_deficiency, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		string(f.ID), itemID, f.Narrative,
		boolToInt(f.IsDeficiency),
		f.CreatedAt.Unix(), f.UpdatedAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("insert finding %s: %w", f.ID, err)
	}

	for _, photo := range f.Photos {
		_, err := tx.Exec(`
			INSERT INTO photos (id, finding_id, storage_path, mime_type, captured_at, uploaded)
			VALUES (?, ?, ?, ?, ?, 0)
		`,
			string(photo.ID), string(f.ID), photo.StoragePath, photo.MimeType, photo.CapturedAt.Unix(),
		)
		if err != nil {
			return fmt.Errorf("insert photo %s: %w", photo.ID, err)
		}
	}
	return nil
}

// FindByID loads a complete Inspection aggregate from SQLite.
func (r *InspectionRepository) FindByID(ctx context.Context, id domain.InspectionID) (*domain.Inspection, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, appointment_id, inspector_id, status,
		       weather, temperature_f, attendees, year_built, structure_type,
		       started_at, completed_at
		FROM inspections WHERE id = ?
	`, string(id))

	insp, err := scanInspectionHeader(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInspectionNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := loadSystems(ctx, r.db, insp); err != nil {
		return nil, err
	}
	return insp, nil
}

// FindByAppointmentID loads a complete Inspection for the given appointment.
func (r *InspectionRepository) FindByAppointmentID(ctx context.Context, appointmentID domain.AppointmentID) (*domain.Inspection, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, appointment_id, inspector_id, status,
		       weather, temperature_f, attendees, year_built, structure_type,
		       started_at, completed_at
		FROM inspections WHERE appointment_id = ?
	`, string(appointmentID))

	insp, err := scanInspectionHeader(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInspectionNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := loadSystems(ctx, r.db, insp); err != nil {
		return nil, err
	}
	return insp, nil
}

// FindByInspector returns inspections for an inspector with optional filters.
func (r *InspectionRepository) FindByInspector(
	ctx context.Context,
	inspectorID domain.InspectorID,
	filter domain.InspectionFilter,
) ([]*domain.Inspection, error) {
	query := `
		SELECT id, appointment_id, inspector_id, status,
		       weather, temperature_f, attendees, year_built, structure_type,
		       started_at, completed_at
		FROM inspections WHERE inspector_id = ?`
	args := []any{string(inspectorID)}

	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, string(*filter.Status))
	}
	if filter.FromDate != nil {
		query += ` AND started_at >= ?`
		args = append(args, filter.FromDate.Unix())
	}
	if filter.ToDate != nil {
		query += ` AND started_at <= ?`
		args = append(args, filter.ToDate.Unix())
	}

	query += ` ORDER BY started_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inspections []*domain.Inspection
	for rows.Next() {
		insp, err := scanInspectionRow(rows)
		if err != nil {
			return nil, err
		}
		if err := loadSystems(ctx, r.db, insp); err != nil {
			return nil, err
		}
		inspections = append(inspections, insp)
	}
	return inspections, rows.Err()
}

// Delete removes an inspection and all its children (cascade).
func (r *InspectionRepository) Delete(ctx context.Context, id domain.InspectionID) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM inspections WHERE id = ?`, string(id)); err != nil {
			return err
		}
		return r.recorder.Record(ctx, tx, "inspections", string(id), "delete", nil)
	})
}

// ── Scan helpers ─────────────────────────────────────────────────────────────

func scanInspectionHeader(row *sql.Row) (*domain.Inspection, error) {
	var (
		id, appointmentID, inspectorID, status string
		weather, attendeesJSON, structureType  string
		temperatureF, yearBuilt                int
		startedAt                              int64
		completedAt                            sql.NullInt64
	)
	err := row.Scan(
		&id, &appointmentID, &inspectorID, &status,
		&weather, &temperatureF, &attendeesJSON, &yearBuilt, &structureType,
		&startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan inspection: %w", err)
	}
	return buildInspection(id, appointmentID, inspectorID, status,
		weather, temperatureF, attendeesJSON, yearBuilt, structureType,
		startedAt, completedAt)
}

func scanInspectionRow(rows *sql.Rows) (*domain.Inspection, error) {
	var (
		id, appointmentID, inspectorID, status string
		weather, attendeesJSON, structureType  string
		temperatureF, yearBuilt                int
		startedAt                              int64
		completedAt                            sql.NullInt64
	)
	err := rows.Scan(
		&id, &appointmentID, &inspectorID, &status,
		&weather, &temperatureF, &attendeesJSON, &yearBuilt, &structureType,
		&startedAt, &completedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan inspection row: %w", err)
	}
	return buildInspection(id, appointmentID, inspectorID, status,
		weather, temperatureF, attendeesJSON, yearBuilt, structureType,
		startedAt, completedAt)
}

func buildInspection(
	id, appointmentID, inspectorID, status string,
	weather string, temperatureF int, attendeesJSON string,
	yearBuilt int, structureType string,
	startedAt int64, completedAt sql.NullInt64,
) (*domain.Inspection, error) {
	var attendees []string
	if err := json.Unmarshal([]byte(attendeesJSON), &attendees); err != nil {
		attendees = []string{}
	}

	var completedAtPtr *time.Time
	if completedAt.Valid {
		t := time.Unix(completedAt.Int64, 0).UTC()
		completedAtPtr = &t
	}

	return &domain.Inspection{
		ID:            domain.InspectionID(id),
		AppointmentID: domain.AppointmentID(appointmentID),
		InspectorID:   domain.InspectorID(inspectorID),
		Status:        domain.InspectionStatus(status),
		Header: domain.InspectionHeader{
			Weather:       weather,
			TemperatureF:  temperatureF,
			Attendees:     attendees,
			YearBuilt:     yearBuilt,
			StructureType: structureType,
		},
		Systems:     make(map[domain.SystemType]*domain.SystemSection),
		StartedAt:   time.Unix(startedAt, 0).UTC(),
		CompletedAt: completedAtPtr,
	}, nil
}

// loadSystems populates insp.Systems from the database.
func loadSystems(ctx context.Context, database *db.DB, insp *domain.Inspection) error {
	rows, err := database.QueryContext(ctx, `
		SELECT id, system_type, inspector_notes, updated_at
		FROM system_sections WHERE inspection_id = ?
	`, string(insp.ID))
	if err != nil {
		return fmt.Errorf("query system_sections: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sectionID, systemType, inspectorNotes string
		var updatedAt int64
		if err := rows.Scan(&sectionID, &systemType, &inspectorNotes, &updatedAt); err != nil {
			return fmt.Errorf("scan system_section: %w", err)
		}

		section := &domain.SystemSection{
			ID:             sectionID,
			SystemType:     domain.SystemType(systemType),
			Descriptions:   make(map[domain.DescriptionKey]string),
			InspectorNotes: inspectorNotes,
			UpdatedAt:      time.Unix(updatedAt, 0).UTC(),
		}

		if err := loadDescriptions(ctx, database, section); err != nil {
			return err
		}
		if err := loadItems(ctx, database, section); err != nil {
			return err
		}

		insp.Systems[domain.SystemType(systemType)] = section
	}
	return rows.Err()
}

func loadDescriptions(ctx context.Context, database *db.DB, section *domain.SystemSection) error {
	rows, err := database.QueryContext(ctx,
		`SELECT description_key, value FROM system_descriptions WHERE system_section_id = ?`,
		section.ID,
	)
	if err != nil {
		return fmt.Errorf("query descriptions for section %s: %w", section.SystemType, err)
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return fmt.Errorf("scan description: %w", err)
		}
		section.Descriptions[domain.DescriptionKey(key)] = value
	}
	return rows.Err()
}

func loadItems(ctx context.Context, database *db.DB, section *domain.SystemSection) error {
	rows, err := database.QueryContext(ctx, `
		SELECT id, item_key, label, status, not_inspected_reason, updated_at
		FROM inspection_items WHERE system_section_id = ?
		ORDER BY rowid ASC
	`, section.ID)
	if err != nil {
		return fmt.Errorf("query items for section %s: %w", section.SystemType, err)
	}
	defer rows.Close()

	for rows.Next() {
		var itemID, itemKey, label, status, niReason string
		var updatedAt int64
		if err := rows.Scan(&itemID, &itemKey, &label, &status, &niReason, &updatedAt); err != nil {
			return fmt.Errorf("scan item: %w", err)
		}

		item := domain.InspectionItem{
			ID:                 itemID,
			ItemKey:            domain.ItemKey(itemKey),
			Label:              label,
			Status:             domain.ItemStatus(status),
			NotInspectedReason: niReason,
			UpdatedAt:          time.Unix(updatedAt, 0).UTC(),
		}

		if err := loadFindings(ctx, database, &item); err != nil {
			return err
		}

		section.Items = append(section.Items, item)
	}
	return rows.Err()
}

func loadFindings(ctx context.Context, database *db.DB, item *domain.InspectionItem) error {
	rows, err := database.QueryContext(ctx, `
		SELECT id, narrative, is_deficiency, created_at, updated_at
		FROM findings WHERE inspection_item_id = ?
		ORDER BY created_at ASC
	`, item.ID)
	if err != nil {
		return fmt.Errorf("query findings for item %s: %w", item.ItemKey, err)
	}
	defer rows.Close()

	for rows.Next() {
		var fid, narrative string
		var isDeficiency int
		var createdAt, updatedAt int64
		if err := rows.Scan(&fid, &narrative, &isDeficiency, &createdAt, &updatedAt); err != nil {
			return fmt.Errorf("scan finding: %w", err)
		}

		f := domain.Finding{
			ID:           domain.FindingID(fid),
			Narrative:    narrative,
			IsDeficiency: isDeficiency != 0,
			CreatedAt:    time.Unix(createdAt, 0).UTC(),
			UpdatedAt:    time.Unix(updatedAt, 0).UTC(),
		}

		if err := loadPhotos(ctx, database, &f); err != nil {
			return err
		}

		item.Findings = append(item.Findings, f)
	}
	return rows.Err()
}

func loadPhotos(ctx context.Context, database *db.DB, f *domain.Finding) error {
	rows, err := database.QueryContext(ctx, `
		SELECT id, storage_path, mime_type, captured_at
		FROM photos WHERE finding_id = ?
		ORDER BY captured_at ASC
	`, string(f.ID))
	if err != nil {
		return fmt.Errorf("query photos for finding %s: %w", f.ID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var pid, storagePath, mimeType string
		var capturedAt int64
		if err := rows.Scan(&pid, &storagePath, &mimeType, &capturedAt); err != nil {
			return fmt.Errorf("scan photo: %w", err)
		}
		f.Photos = append(f.Photos, domain.PhotoRef{
			ID:          domain.PhotoID(pid),
			StoragePath: storagePath,
			MimeType:    mimeType,
			CapturedAt:  time.Unix(capturedAt, 0).UTC(),
		})
	}
	return rows.Err()
}

// FindPhotoMeta returns the storage path and MIME type for a photo.
// Used to serve photos without loading the full inspection aggregate.
func (r *InspectionRepository) FindPhotoMeta(ctx context.Context, photoID domain.PhotoID) (storagePath, mimeType string, err error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT storage_path, mime_type FROM photos WHERE id = ?`,
		string(photoID),
	)
	if err = row.Scan(&storagePath, &mimeType); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", domain.ErrPhotoNotFound
		}
		return "", "", fmt.Errorf("scan photo meta: %w", err)
	}
	return storagePath, mimeType, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
