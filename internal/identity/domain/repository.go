package domain

import "context"

// InspectorRepository is the persistence contract for Inspector aggregates.
type InspectorRepository interface {
	Save(ctx context.Context, inspector *Inspector) error
	FindByID(ctx context.Context, id InspectorID) (*Inspector, error)
	FindByEmail(ctx context.Context, email string) (*Inspector, error)
	FindByCompany(ctx context.Context, companyID CompanyID) ([]*Inspector, error)
	Delete(ctx context.Context, id InspectorID) error
}

// CompanyRepository is the persistence contract for Company aggregates.
type CompanyRepository interface {
	Save(ctx context.Context, company *Company) error
	FindByID(ctx context.Context, id CompanyID) (*Company, error)
	Delete(ctx context.Context, id CompanyID) error
}

// ClientRepository is the persistence contract for Client aggregates.
type ClientRepository interface {
	Save(ctx context.Context, client *Client) error
	FindByID(ctx context.Context, id ClientID) (*Client, error)
	FindByCompany(ctx context.Context, companyID CompanyID, filter ClientFilter) ([]*Client, error)
	Delete(ctx context.Context, id ClientID) error
}

type ClientFilter struct {
	Search string // name or email substring match
	Limit  int
	Offset int
}
