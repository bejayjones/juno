package application_test

import (
	"context"
	"time"

	"github.com/bejayjones/juno/internal/identity/domain"
)

// fakeInspectorRepo is an in-memory InspectorRepository for tests.
type fakeInspectorRepo struct {
	byID    map[domain.InspectorID]*domain.Inspector
	byEmail map[string]*domain.Inspector
}

func newFakeInspectorRepo() *fakeInspectorRepo {
	return &fakeInspectorRepo{
		byID:    make(map[domain.InspectorID]*domain.Inspector),
		byEmail: make(map[string]*domain.Inspector),
	}
}

func (r *fakeInspectorRepo) Save(_ context.Context, i *domain.Inspector) error {
	// Remove stale email index if email changed.
	for email, existing := range r.byEmail {
		if existing.ID == i.ID && email != i.Email {
			delete(r.byEmail, email)
		}
	}
	r.byID[i.ID] = i
	r.byEmail[i.Email] = i
	return nil
}

func (r *fakeInspectorRepo) FindByID(_ context.Context, id domain.InspectorID) (*domain.Inspector, error) {
	if i, ok := r.byID[id]; ok {
		return i, nil
	}
	return nil, domain.ErrInspectorNotFound
}

func (r *fakeInspectorRepo) FindByEmail(_ context.Context, email string) (*domain.Inspector, error) {
	if i, ok := r.byEmail[email]; ok {
		return i, nil
	}
	return nil, domain.ErrInspectorNotFound
}

func (r *fakeInspectorRepo) FindByCompany(_ context.Context, companyID domain.CompanyID) ([]*domain.Inspector, error) {
	var result []*domain.Inspector
	for _, i := range r.byID {
		if i.CompanyID == companyID {
			result = append(result, i)
		}
	}
	return result, nil
}

func (r *fakeInspectorRepo) Delete(_ context.Context, id domain.InspectorID) error {
	if i, ok := r.byID[id]; ok {
		delete(r.byEmail, i.Email)
		delete(r.byID, id)
	}
	return nil
}

// fakeCompanyRepo is an in-memory CompanyRepository for tests.
type fakeCompanyRepo struct {
	byID map[domain.CompanyID]*domain.Company
}

func newFakeCompanyRepo() *fakeCompanyRepo {
	return &fakeCompanyRepo{byID: make(map[domain.CompanyID]*domain.Company)}
}

func (r *fakeCompanyRepo) Save(_ context.Context, c *domain.Company) error {
	r.byID[c.ID] = c
	return nil
}

func (r *fakeCompanyRepo) FindByID(_ context.Context, id domain.CompanyID) (*domain.Company, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, domain.ErrCompanyNotFound
}

func (r *fakeCompanyRepo) Delete(_ context.Context, id domain.CompanyID) error {
	delete(r.byID, id)
	return nil
}

// fakeClientRepo is an in-memory ClientRepository for tests.
type fakeClientRepo struct {
	byID map[domain.ClientID]*domain.Client
}

func newFakeClientRepo() *fakeClientRepo {
	return &fakeClientRepo{byID: make(map[domain.ClientID]*domain.Client)}
}

func (r *fakeClientRepo) Save(_ context.Context, c *domain.Client) error {
	r.byID[c.ID] = c
	return nil
}

func (r *fakeClientRepo) FindByID(_ context.Context, id domain.ClientID) (*domain.Client, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, domain.ErrClientNotFound
}

func (r *fakeClientRepo) FindByCompany(_ context.Context, companyID domain.CompanyID, filter domain.ClientFilter) ([]*domain.Client, error) {
	var result []*domain.Client
	for _, c := range r.byID {
		if c.CompanyID == companyID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *fakeClientRepo) Delete(_ context.Context, id domain.ClientID) error {
	delete(r.byID, id)
	return nil
}

// fakeHasher is a no-op PasswordHasher for tests (not cryptographically safe).
type fakeHasher struct{}

func (fakeHasher) Hash(p string) (string, error)      { return "hashed:" + p, nil }
func (fakeHasher) Verify(hash, p string) bool         { return hash == "hashed:"+p }

// fakeTokenIssuer returns a deterministic token for tests.
type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(inspectorID, companyID, role string) (string, time.Time, error) {
	return "test-token", time.Now().Add(24 * time.Hour), nil
}

// fixedClock returns a fixed time for deterministic tests.
type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

var testNow = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
