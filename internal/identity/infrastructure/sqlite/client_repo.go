package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/bejayjones/juno/internal/platform/db"
	"github.com/bejayjones/juno/internal/sync/recorder"
)

type ClientRepository struct {
	db       *db.DB
	recorder *recorder.Recorder
}

func NewClientRepository(database *db.DB) *ClientRepository {
	return &ClientRepository{db: database}
}

// WithRecorder enables sync recording for this repository.
func (r *ClientRepository) WithRecorder(rec *recorder.Recorder) *ClientRepository {
	r.recorder = rec
	return r
}

func (r *ClientRepository) Save(ctx context.Context, client *domain.Client) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO clients
				(id, company_id, first_name, last_name, email, phone, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET
				first_name = excluded.first_name,
				last_name  = excluded.last_name,
				email      = excluded.email,
				phone      = excluded.phone,
				updated_at = excluded.updated_at
		`,
			string(client.ID), string(client.CompanyID),
			client.Name.First, client.Name.Last,
			client.Email, client.Phone,
			client.CreatedAt.Unix(), client.UpdatedAt.Unix(),
		)
		if err != nil {
			return fmt.Errorf("upsert client: %w", err)
		}
		return r.recorder.Record(ctx, tx, "clients", string(client.ID), "upsert", client)
	})
}

func (r *ClientRepository) FindByID(ctx context.Context, id domain.ClientID) (*domain.Client, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, company_id, first_name, last_name, email, phone, created_at, updated_at
		FROM clients WHERE id = ?
	`, string(id))

	client, err := scanClient(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrClientNotFound
	}
	return client, err
}

func (r *ClientRepository) FindByCompany(ctx context.Context, companyID domain.CompanyID, filter domain.ClientFilter) ([]*domain.Client, error) {
	query := `
		SELECT id, company_id, first_name, last_name, email, phone, created_at, updated_at
		FROM clients WHERE company_id = ?`
	args := []any{string(companyID)}

	if filter.Search != "" {
		query += ` AND (first_name LIKE ? OR last_name LIKE ? OR email LIKE ?)`
		s := "%" + filter.Search + "%"
		args = append(args, s, s, s)
	}

	query += ` ORDER BY last_name, first_name`

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

	var clients []*domain.Client
	for rows.Next() {
		var (
			id, companyID_, firstName, lastName, email, phone string
			createdAt, updatedAt                              int64
		)
		if err := rows.Scan(&id, &companyID_, &firstName, &lastName, &email, &phone, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan client: %w", err)
		}
		clients = append(clients, &domain.Client{
			ID:        domain.ClientID(id),
			CompanyID: domain.CompanyID(companyID_),
			Name:      domain.PersonName{First: firstName, Last: lastName},
			Email:     email,
			Phone:     phone,
			CreatedAt: time.Unix(createdAt, 0).UTC(),
			UpdatedAt: time.Unix(updatedAt, 0).UTC(),
		})
	}
	return clients, rows.Err()
}

func (r *ClientRepository) Delete(ctx context.Context, id domain.ClientID) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM clients WHERE id = ?`, string(id)); err != nil {
			return err
		}
		return r.recorder.Record(ctx, tx, "clients", string(id), "delete", nil)
	})
}

func scanClient(row *sql.Row) (*domain.Client, error) {
	var (
		id, companyID, firstName, lastName, email, phone string
		createdAt, updatedAt                              int64
	)
	err := row.Scan(&id, &companyID, &firstName, &lastName, &email, &phone, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan client: %w", err)
	}
	return &domain.Client{
		ID:        domain.ClientID(id),
		CompanyID: domain.CompanyID(companyID),
		Name:      domain.PersonName{First: firstName, Last: lastName},
		Email:     email,
		Phone:     phone,
		CreatedAt: time.Unix(createdAt, 0).UTC(),
		UpdatedAt: time.Unix(updatedAt, 0).UTC(),
	}, nil
}
