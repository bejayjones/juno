package application

import (
	"context"
	"fmt"

	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/bejayjones/juno/pkg/clock"
)

// CompanyService handles reads and updates for Company aggregates.
type CompanyService struct {
	companies domain.CompanyRepository
	clock     clock.Clock
}

func NewCompanyService(companies domain.CompanyRepository, clk clock.Clock) *CompanyService {
	return &CompanyService{companies: companies, clock: clk}
}

// GetByID returns a company by ID.
func (s *CompanyService) GetByID(ctx context.Context, id domain.CompanyID) (CompanyView, error) {
	company, err := s.companies.FindByID(ctx, id)
	if err != nil {
		return CompanyView{}, err
	}
	return toCompanyView(company), nil
}

// UpdateCompanyInput contains the mutable fields for a company profile.
type UpdateCompanyInput struct {
	Name    string
	Street  string
	City    string
	State   string
	Zip     string
	Country string
	Phone   string
	Email   string
}

// Update changes mutable company fields.
func (s *CompanyService) Update(ctx context.Context, id domain.CompanyID, in UpdateCompanyInput) (CompanyView, error) {
	company, err := s.companies.FindByID(ctx, id)
	if err != nil {
		return CompanyView{}, err
	}

	company.UpdateProfile(
		in.Name,
		domain.Address{Street: in.Street, City: in.City, State: in.State, Zip: in.Zip, Country: in.Country},
		in.Phone, in.Email,
		s.clock.Now(),
	)
	if err := s.companies.Save(ctx, company); err != nil {
		return CompanyView{}, fmt.Errorf("save company: %w", err)
	}
	return toCompanyView(company), nil
}
