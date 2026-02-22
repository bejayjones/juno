// Package sqlite provides the SQLite-backed repository for the reporting context.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/bejayjones/juno/internal/platform/db"
	"github.com/bejayjones/juno/internal/reporting/domain"
	"github.com/bejayjones/juno/internal/sync/recorder"
)

// ReportRepository persists Report aggregates in SQLite.
type ReportRepository struct {
	db       *db.DB
	recorder *recorder.Recorder
}

func NewReportRepository(database *db.DB) *ReportRepository {
	return &ReportRepository{db: database}
}

// WithRecorder enables sync recording for this repository.
func (r *ReportRepository) WithRecorder(rec *recorder.Recorder) *ReportRepository {
	r.recorder = rec
	return r
}

// Save upserts the Report and replaces its Deliveries within a single transaction.
func (r *ReportRepository) Save(ctx context.Context, report *domain.Report) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		var genAt *int64
		if report.GeneratedAt != nil {
			v := report.GeneratedAt.Unix()
			genAt = &v
		}

		_, err := tx.ExecContext(ctx, `
			INSERT INTO reports (id, inspection_id, inspector_id, status, pdf_storage_path, generated_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				status           = excluded.status,
				pdf_storage_path = excluded.pdf_storage_path,
				generated_at     = excluded.generated_at,
				updated_at       = excluded.updated_at`,
			string(report.ID),
			string(report.InspectionID),
			string(report.InspectorID),
			string(report.Status),
			report.PDFStoragePath,
			genAt,
			report.CreatedAt.Unix(),
			report.UpdatedAt.Unix(),
		)
		if err != nil {
			return err
		}

		// Delete + reinsert deliveries to keep them in sync.
		if _, err := tx.ExecContext(ctx, `DELETE FROM deliveries WHERE report_id = ?`, string(report.ID)); err != nil {
			return err
		}
		for _, d := range report.Deliveries {
			var sentAt *int64
			if d.SentAt != nil {
				v := d.SentAt.Unix()
				sentAt = &v
			}
			_, err := tx.ExecContext(ctx, `
				INSERT INTO deliveries (id, report_id, recipient_email, status, attempts, sent_at, failure_reason, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				string(d.ID),
				string(report.ID),
				d.RecipientEmail,
				string(d.Status),
				d.Attempts,
				sentAt,
				d.FailureReason,
				d.CreatedAt.Unix(),
				d.UpdatedAt.Unix(),
			)
			if err != nil {
				return err
			}
		}
		return r.recorder.Record(ctx, tx, "reports", string(report.ID), "upsert", report)
	})
}

func (r *ReportRepository) FindByID(ctx context.Context, id domain.ReportID) (*domain.Report, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, inspection_id, inspector_id, status, pdf_storage_path, generated_at, created_at, updated_at
		FROM reports WHERE id = ?`, string(id))

	report, err := scanReport(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReportNotFound
		}
		return nil, err
	}
	return report, r.loadDeliveries(ctx, report)
}

func (r *ReportRepository) FindByInspection(ctx context.Context, inspectionID domain.InspectionID) (*domain.Report, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, inspection_id, inspector_id, status, pdf_storage_path, generated_at, created_at, updated_at
		FROM reports WHERE inspection_id = ?`, string(inspectionID))

	report, err := scanReport(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReportNotFound
		}
		return nil, err
	}
	return report, r.loadDeliveries(ctx, report)
}

func (r *ReportRepository) FindByInspector(ctx context.Context, inspectorID domain.InspectorID, filter domain.ReportFilter) ([]*domain.Report, error) {
	query := `
		SELECT id, inspection_id, inspector_id, status, pdf_storage_path, generated_at, created_at, updated_at
		FROM reports WHERE inspector_id = ?`
	args := []any{string(inspectorID)}

	if filter.Status != nil {
		query += " AND status = ?"
		args = append(args, string(*filter.Status))
	}
	query += " ORDER BY created_at DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// Materialize report headers before loading deliveries to avoid nested
	// query deadlock with SQLite MaxOpenConns(1).
	var reports []*domain.Report
	for rows.Next() {
		report, err := scanReport(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		reports = append(reports, report)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, report := range reports {
		if err := r.loadDeliveries(ctx, report); err != nil {
			return nil, err
		}
	}
	return reports, nil
}

func (r *ReportRepository) Delete(ctx context.Context, id domain.ReportID) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `DELETE FROM reports WHERE id = ?`, string(id))
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		if n == 0 {
			return domain.ErrReportNotFound
		}
		return r.recorder.Record(ctx, tx, "reports", string(id), "delete", nil)
	})
}

// scanner covers both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanReport(s scanner) (*domain.Report, error) {
	var (
		r          domain.Report
		idStr      string
		inspID     string
		instID     string
		status     string
		genAt      sql.NullInt64
		createdAt  int64
		updatedAt  int64
	)
	err := s.Scan(&idStr, &inspID, &instID, &status, &r.PDFStoragePath, &genAt, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	r.ID = domain.ReportID(idStr)
	r.InspectionID = domain.InspectionID(inspID)
	r.InspectorID = domain.InspectorID(instID)
	r.Status = domain.ReportStatus(status)
	r.CreatedAt = time.Unix(createdAt, 0)
	r.UpdatedAt = time.Unix(updatedAt, 0)
	if genAt.Valid {
		t := time.Unix(genAt.Int64, 0)
		r.GeneratedAt = &t
	}
	return &r, nil
}

func (r *ReportRepository) loadDeliveries(ctx context.Context, report *domain.Report) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, recipient_email, status, attempts, sent_at, failure_reason, created_at, updated_at
		FROM deliveries WHERE report_id = ? ORDER BY created_at`,
		string(report.ID))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			d         domain.Delivery
			idStr     string
			status    string
			sentAt    sql.NullInt64
			createdAt int64
			updatedAt int64
		)
		err := rows.Scan(&idStr, &d.RecipientEmail, &status, &d.Attempts, &sentAt, &d.FailureReason, &createdAt, &updatedAt)
		if err != nil {
			return err
		}
		d.ID = domain.DeliveryID(idStr)
		d.Status = domain.DeliveryStatus(status)
		d.CreatedAt = time.Unix(createdAt, 0)
		d.UpdatedAt = time.Unix(updatedAt, 0)
		if sentAt.Valid {
			t := time.Unix(sentAt.Int64, 0)
			d.SentAt = &t
		}
		report.Deliveries = append(report.Deliveries, d)
	}
	return rows.Err()
}
