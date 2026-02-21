package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/id"
)

// InspectorService handles inspector registration, login, and profile management.
type InspectorService struct {
	inspectors domain.InspectorRepository
	companies  domain.CompanyRepository
	hasher     PasswordHasher
	tokens     TokenIssuer
	clock      clock.Clock
}

func NewInspectorService(
	inspectors domain.InspectorRepository,
	companies domain.CompanyRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	clk clock.Clock,
) *InspectorService {
	return &InspectorService{
		inspectors: inspectors,
		companies:  companies,
		hasher:     hasher,
		tokens:     tokens,
		clock:      clk,
	}
}

// RegisterInput contains fields needed to create a new inspector account.
// Either CompanyID (join existing) or CompanyName (create new) must be provided.
type RegisterInput struct {
	FirstName   string
	LastName    string
	Email       string
	Password    string
	CompanyID   string // omit to create a new company
	CompanyName string // required when CompanyID is empty
}

// RegisterOutput contains the newly created inspector, their company, and an auth token.
type RegisterOutput struct {
	Inspector InspectorView `json:"inspector"`
	Company   CompanyView   `json:"company"`
	Token     string        `json:"token"`
	ExpiresAt int64         `json:"expires_at"`
}

// Register creates a new inspector. If CompanyID is empty a new company is created
// and the inspector is assigned the "owner" role. Otherwise the inspector joins
// the named company as a "member".
func (s *InspectorService) Register(ctx context.Context, in RegisterInput) (RegisterOutput, error) {
	_, err := s.inspectors.FindByEmail(ctx, in.Email)
	if err == nil {
		return RegisterOutput{}, domain.ErrEmailTaken
	}
	if !errors.Is(err, domain.ErrInspectorNotFound) {
		return RegisterOutput{}, fmt.Errorf("check email: %w", err)
	}

	now := s.clock.Now()

	var company *domain.Company
	var role domain.InspectorRole

	if in.CompanyID != "" {
		company, err = s.companies.FindByID(ctx, domain.CompanyID(in.CompanyID))
		if err != nil {
			return RegisterOutput{}, fmt.Errorf("find company: %w", err)
		}
		role = domain.RoleMember
	} else {
		if in.CompanyName == "" {
			return RegisterOutput{}, fmt.Errorf("company_name is required when company_id is not provided")
		}
		company = domain.NewCompany(
			domain.CompanyID(id.New()),
			in.CompanyName,
			domain.Address{},
			"", "",
			now,
		)
		if err := s.companies.Save(ctx, company); err != nil {
			return RegisterOutput{}, fmt.Errorf("save company: %w", err)
		}
		role = domain.RoleOwner
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("hash password: %w", err)
	}

	inspector, err := domain.NewInspector(
		domain.InspectorID(id.New()),
		company.ID,
		domain.PersonName{First: in.FirstName, Last: in.LastName},
		in.Email,
		hash,
		role,
		now,
	)
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("create inspector: %w", err)
	}
	if err := s.inspectors.Save(ctx, inspector); err != nil {
		return RegisterOutput{}, fmt.Errorf("save inspector: %w", err)
	}

	token, exp, err := s.tokens.Issue(string(inspector.ID), string(company.ID), string(role))
	if err != nil {
		return RegisterOutput{}, fmt.Errorf("issue token: %w", err)
	}

	return RegisterOutput{
		Inspector: toInspectorView(inspector),
		Company:   toCompanyView(company),
		Token:     token,
		ExpiresAt: exp.Unix(),
	}, nil
}

// LoginOutput contains the authenticated inspector and a fresh token.
type LoginOutput struct {
	Inspector InspectorView `json:"inspector"`
	Token     string        `json:"token"`
	ExpiresAt int64         `json:"expires_at"`
}

// Login verifies credentials and returns a token on success.
func (s *InspectorService) Login(ctx context.Context, email, password string) (LoginOutput, error) {
	inspector, err := s.inspectors.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrInspectorNotFound) {
			return LoginOutput{}, domain.ErrInspectorNotFound
		}
		return LoginOutput{}, fmt.Errorf("find inspector: %w", err)
	}

	if !s.hasher.Verify(inspector.PasswordHash, password) {
		return LoginOutput{}, domain.ErrInspectorNotFound // intentionally vague
	}

	token, exp, err := s.tokens.Issue(
		string(inspector.ID),
		string(inspector.CompanyID),
		string(inspector.Role),
	)
	if err != nil {
		return LoginOutput{}, fmt.Errorf("issue token: %w", err)
	}

	return LoginOutput{
		Inspector: toInspectorView(inspector),
		Token:     token,
		ExpiresAt: exp.Unix(),
	}, nil
}

// GetByID returns an inspector by their ID.
func (s *InspectorService) GetByID(ctx context.Context, inspectorID domain.InspectorID) (InspectorView, error) {
	inspector, err := s.inspectors.FindByID(ctx, inspectorID)
	if err != nil {
		return InspectorView{}, err
	}
	return toInspectorView(inspector), nil
}

// UpdateProfileInput contains the mutable fields for an inspector's profile.
type UpdateProfileInput struct {
	FirstName string
	LastName  string
	Email     string
}

// UpdateProfile changes name and email for the given inspector.
func (s *InspectorService) UpdateProfile(ctx context.Context, inspectorID domain.InspectorID, in UpdateProfileInput) (InspectorView, error) {
	inspector, err := s.inspectors.FindByID(ctx, inspectorID)
	if err != nil {
		return InspectorView{}, err
	}

	if in.Email != inspector.Email {
		_, err := s.inspectors.FindByEmail(ctx, in.Email)
		if err == nil {
			return InspectorView{}, domain.ErrEmailTaken
		}
		if !errors.Is(err, domain.ErrInspectorNotFound) {
			return InspectorView{}, fmt.Errorf("check email: %w", err)
		}
	}

	inspector.UpdateProfile(
		domain.PersonName{First: in.FirstName, Last: in.LastName},
		in.Email,
		s.clock.Now(),
	)
	if err := s.inspectors.Save(ctx, inspector); err != nil {
		return InspectorView{}, fmt.Errorf("save inspector: %w", err)
	}
	return toInspectorView(inspector), nil
}

// SetLicenseInput contains a state license to add or update.
type SetLicenseInput struct {
	State  string
	Number string
}

// SetLicense adds or updates a state license for the given inspector.
func (s *InspectorService) SetLicense(ctx context.Context, inspectorID domain.InspectorID, in SetLicenseInput) (InspectorView, error) {
	inspector, err := s.inspectors.FindByID(ctx, inspectorID)
	if err != nil {
		return InspectorView{}, err
	}

	inspector.SetLicense(domain.LicenseNumber{State: in.State, Number: in.Number}, s.clock.Now())
	if err := s.inspectors.Save(ctx, inspector); err != nil {
		return InspectorView{}, fmt.Errorf("save inspector: %w", err)
	}
	return toInspectorView(inspector), nil
}

// ListByCompany returns all inspectors belonging to a company.
func (s *InspectorService) ListByCompany(ctx context.Context, companyID domain.CompanyID) ([]InspectorView, error) {
	inspectors, err := s.inspectors.FindByCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("list inspectors: %w", err)
	}
	views := make([]InspectorView, len(inspectors))
	for i, insp := range inspectors {
		views[i] = toInspectorView(insp)
	}
	return views, nil
}
