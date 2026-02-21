// Package sqlite provides SQLite-backed repository implementations for the
// identity bounded context.
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

type CompanyRepository struct {
	db *db.DB
}

func NewCompanyRepository(database *db.DB) *CompanyRepository {
	return &CompanyRepository{db: database}
}

func (r *CompanyRepository) Save(ctx context.Context, c *domain.Company) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO companies
			(id, name, street, city, state, zip, country, phone, email, logo_storage_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name              = excluded.name,
			street            = excluded.street,
			city              = excluded.city,
			state             = excluded.state,
			zip               = excluded.zip,
			country           = excluded.country,
			phone             = excluded.phone,
			email             = excluded.email,
			logo_storage_path = excluded.logo_storage_path,
			updated_at        = excluded.updated_at
	`,
		string(c.ID), c.Name,
		c.Address.Street, c.Address.City, c.Address.State, c.Address.Zip, c.Address.Country,
		c.Phone, c.Email, c.LogoStoragePath,
		c.CreatedAt.Unix(), c.UpdatedAt.Unix(),
	)
	return err
}

func (r *CompanyRepository) FindByID(ctx context.Context, id domain.CompanyID) (*domain.Company, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, street, city, state, zip, country, phone, email, logo_storage_path, created_at, updated_at
		FROM companies WHERE id = ?
	`, string(id))

	c, err := scanCompany(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCompanyNotFound
	}
	return c, err
}

func (r *CompanyRepository) Delete(ctx context.Context, id domain.CompanyID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM companies WHERE id = ?`, string(id))
	return err
}

func scanCompany(row *sql.Row) (*domain.Company, error) {
	var (
		id, name                         string
		street, city, state, zip, country string
		phone, email, logo               string
		createdAt, updatedAt             int64
	)
	err := row.Scan(&id, &name, &street, &city, &state, &zip, &country, &phone, &email, &logo, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan company: %w", err)
	}
	return &domain.Company{
		ID:              domain.CompanyID(id),
		Name:            name,
		Address:         domain.Address{Street: street, City: city, State: state, Zip: zip, Country: country},
		Phone:           phone,
		Email:           email,
		LogoStoragePath: logo,
		CreatedAt:       time.Unix(createdAt, 0).UTC(),
		UpdatedAt:       time.Unix(updatedAt, 0).UTC(),
	}, nil
}
