package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/bejayjones/juno/internal/platform/db"
)

type InspectorRepository struct {
	db *db.DB
}

func NewInspectorRepository(database *db.DB) *InspectorRepository {
	return &InspectorRepository{db: database}
}

func (r *InspectorRepository) Save(ctx context.Context, inspector *domain.Inspector) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO inspectors
				(id, company_id, first_name, last_name, email, password_hash, role, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				first_name    = excluded.first_name,
				last_name     = excluded.last_name,
				email         = excluded.email,
				password_hash = excluded.password_hash,
				role          = excluded.role,
				updated_at    = excluded.updated_at
		`,
			string(inspector.ID), string(inspector.CompanyID),
			inspector.Name.First, inspector.Name.Last,
			inspector.Email, inspector.PasswordHash, string(inspector.Role),
			inspector.CreatedAt.Unix(), inspector.UpdatedAt.Unix(),
		)
		if err != nil {
			return fmt.Errorf("upsert inspector: %w", err)
		}

		// Sync licenses: replace all with current set.
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM inspector_licenses WHERE inspector_id = ?`, string(inspector.ID),
		); err != nil {
			return fmt.Errorf("delete licenses: %w", err)
		}
		for _, lic := range inspector.LicenseNumbers {
			if _, err := tx.ExecContext(ctx,
				`INSERT INTO inspector_licenses (inspector_id, state, license_number) VALUES (?, ?, ?)`,
				string(inspector.ID), lic.State, lic.Number,
			); err != nil {
				return fmt.Errorf("insert license: %w", err)
			}
		}
		return nil
	})
}

func (r *InspectorRepository) FindByID(ctx context.Context, id domain.InspectorID) (*domain.Inspector, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, first_name, last_name, email, password_hash, role, created_at, updated_at
		FROM inspectors WHERE id = ?
	`, string(id))

	inspector, err := scanInspector(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInspectorNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.loadLicenses(ctx, inspector)
}

func (r *InspectorRepository) FindByEmail(ctx context.Context, email string) (*domain.Inspector, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, first_name, last_name, email, password_hash, role, created_at, updated_at
		FROM inspectors WHERE email = ?
	`, email)

	inspector, err := scanInspector(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInspectorNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.loadLicenses(ctx, inspector)
}

func (r *InspectorRepository) FindByCompany(ctx context.Context, companyID domain.CompanyID) ([]*domain.Inspector, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, company_id, first_name, last_name, email, password_hash, role, created_at, updated_at
		FROM inspectors WHERE company_id = ? ORDER BY last_name, first_name
	`, string(companyID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inspectors []*domain.Inspector
	for rows.Next() {
		inspector, err := scanInspectorRow(rows)
		if err != nil {
			return nil, err
		}
		inspector, err = r.loadLicenses(ctx, inspector)
		if err != nil {
			return nil, err
		}
		inspectors = append(inspectors, inspector)
	}
	return inspectors, rows.Err()
}

func (r *InspectorRepository) Delete(ctx context.Context, id domain.InspectorID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM inspectors WHERE id = ?`, string(id))
	return err
}

func (r *InspectorRepository) loadLicenses(ctx context.Context, inspector *domain.Inspector) (*domain.Inspector, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT state, license_number FROM inspector_licenses WHERE inspector_id = ?`,
		string(inspector.ID),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var lic domain.LicenseNumber
		if err := rows.Scan(&lic.State, &lic.Number); err != nil {
			return nil, err
		}
		inspector.LicenseNumbers = append(inspector.LicenseNumbers, lic)
	}
	return inspector, rows.Err()
}

func scanInspector(row *sql.Row) (*domain.Inspector, error) {
	var (
		id, companyID, firstName, lastName string
		email, passwordHash, role          string
		createdAt, updatedAt               int64
	)
	err := row.Scan(&id, &companyID, &firstName, &lastName, &email, &passwordHash, &role, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return &domain.Inspector{
		ID:           domain.InspectorID(id),
		CompanyID:    domain.CompanyID(companyID),
		Name:         domain.PersonName{First: firstName, Last: lastName},
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.InspectorRole(role),
		CreatedAt:    time.Unix(createdAt, 0).UTC(),
		UpdatedAt:    time.Unix(updatedAt, 0).UTC(),
	}, nil
}

func scanInspectorRow(rows *sql.Rows) (*domain.Inspector, error) {
	var (
		id, companyID, firstName, lastName string
		email, passwordHash, role          string
		createdAt, updatedAt               int64
	)
	err := rows.Scan(&id, &companyID, &firstName, &lastName, &email, &passwordHash, &role, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan inspector: %w", err)
	}
	return &domain.Inspector{
		ID:           domain.InspectorID(id),
		CompanyID:    domain.CompanyID(companyID),
		Name:         domain.PersonName{First: firstName, Last: lastName},
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.InspectorRole(role),
		CreatedAt:    time.Unix(createdAt, 0).UTC(),
		UpdatedAt:    time.Unix(updatedAt, 0).UTC(),
	}, nil
}
