// Package sqlite provides the SQLite-backed repository for the scheduling context.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bejayjones/juno/internal/platform/db"
	"github.com/bejayjones/juno/internal/scheduling/domain"
	"github.com/bejayjones/juno/internal/sync/recorder"
)

// AppointmentRepository persists Appointment aggregates in SQLite.
type AppointmentRepository struct {
	db       *db.DB
	recorder *recorder.Recorder
}

func NewAppointmentRepository(database *db.DB) *AppointmentRepository {
	return &AppointmentRepository{db: database}
}

// WithRecorder enables sync recording for this repository.
func (r *AppointmentRepository) WithRecorder(rec *recorder.Recorder) *AppointmentRepository {
	r.recorder = rec
	return r
}

func (r *AppointmentRepository) Save(ctx context.Context, a *domain.Appointment) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO appointments
				(id, inspector_id, client_id, street, city, state, zip, country,
				 scheduled_at, estimated_duration_min, status, notes, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				street                 = excluded.street,
				city                   = excluded.city,
				state                  = excluded.state,
				zip                    = excluded.zip,
				country                = excluded.country,
				scheduled_at           = excluded.scheduled_at,
				estimated_duration_min = excluded.estimated_duration_min,
				status                 = excluded.status,
				notes                  = excluded.notes,
				updated_at             = excluded.updated_at
		`,
			string(a.ID), string(a.InspectorID), string(a.ClientID),
			a.Property.Street, a.Property.City, a.Property.State, a.Property.Zip, a.Property.Country,
			a.ScheduledAt.Unix(), a.EstimatedDurationMin,
			string(a.Status), a.Notes,
			a.CreatedAt.Unix(), a.UpdatedAt.Unix(),
		)
		if err != nil {
			return fmt.Errorf("upsert appointment: %w", err)
		}
		return r.recorder.Record(ctx, tx, "appointments", string(a.ID), "upsert", a)
	})
}

func (r *AppointmentRepository) FindByID(ctx context.Context, id domain.AppointmentID) (*domain.Appointment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, inspector_id, client_id, street, city, state, zip, country,
		       scheduled_at, estimated_duration_min, status, notes, created_at, updated_at
		FROM appointments WHERE id = ?
	`, string(id))

	a, err := scanAppointment(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAppointmentNotFound
	}
	return a, err
}

func (r *AppointmentRepository) FindByInspector(ctx context.Context, inspectorID domain.InspectorID, filter domain.AppointmentFilter) ([]*domain.Appointment, error) {
	query := `
		SELECT id, inspector_id, client_id, street, city, state, zip, country,
		       scheduled_at, estimated_duration_min, status, notes, created_at, updated_at
		FROM appointments WHERE inspector_id = ?`
	args := []any{string(inspectorID)}

	if filter.Status != nil {
		query += ` AND status = ?`
		args = append(args, string(*filter.Status))
	}
	if filter.FromDate != nil {
		query += ` AND scheduled_at >= ?`
		args = append(args, filter.FromDate.Unix())
	}
	if filter.ToDate != nil {
		query += ` AND scheduled_at <= ?`
		args = append(args, filter.ToDate.Unix())
	}

	query += ` ORDER BY scheduled_at ASC`

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

	var appointments []*domain.Appointment
	for rows.Next() {
		a, err := scanAppointmentRow(rows)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, rows.Err()
}

func (r *AppointmentRepository) Delete(ctx context.Context, id domain.AppointmentID) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM appointments WHERE id = ?`, string(id)); err != nil {
			return err
		}
		return r.recorder.Record(ctx, tx, "appointments", string(id), "delete", nil)
	})
}

func scanAppointment(row *sql.Row) (*domain.Appointment, error) {
	var (
		id, inspectorID, clientID         string
		street, city, state, zip, country string
		scheduledAt                       int64
		durationMin                       int
		status, notes                     string
		createdAt, updatedAt              int64
	)
	err := row.Scan(
		&id, &inspectorID, &clientID,
		&street, &city, &state, &zip, &country,
		&scheduledAt, &durationMin, &status, &notes,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan appointment: %w", err)
	}
	return buildAppointment(id, inspectorID, clientID, street, city, state, zip, country,
		scheduledAt, durationMin, status, notes, createdAt, updatedAt), nil
}

func scanAppointmentRow(rows *sql.Rows) (*domain.Appointment, error) {
	var (
		id, inspectorID, clientID         string
		street, city, state, zip, country string
		scheduledAt                       int64
		durationMin                       int
		status, notes                     string
		createdAt, updatedAt              int64
	)
	err := rows.Scan(
		&id, &inspectorID, &clientID,
		&street, &city, &state, &zip, &country,
		&scheduledAt, &durationMin, &status, &notes,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan appointment: %w", err)
	}
	return buildAppointment(id, inspectorID, clientID, street, city, state, zip, country,
		scheduledAt, durationMin, status, notes, createdAt, updatedAt), nil
}

func buildAppointment(id, inspectorID, clientID, street, city, state, zip, country string,
	scheduledAt int64, durationMin int, status, notes string, createdAt, updatedAt int64) *domain.Appointment {
	apptStatus := domain.AppointmentStatus(status)
	return &domain.Appointment{
		ID:          domain.AppointmentID(id),
		InspectorID: domain.InspectorID(inspectorID),
		ClientID:    domain.ClientID(clientID),
		Property: domain.PropertyAddress{
			Street: street, City: city, State: state, Zip: zip, Country: country,
		},
		ScheduledAt:          time.Unix(scheduledAt, 0).UTC(),
		EstimatedDurationMin: durationMin,
		Status:               apptStatus,
		Notes:                notes,
		CreatedAt:            time.Unix(createdAt, 0).UTC(),
		UpdatedAt:            time.Unix(updatedAt, 0).UTC(),
	}
}
