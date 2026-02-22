package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bejayjones/juno/api/rest"
	"github.com/bejayjones/juno/api/rest/middleware"
	identityapp "github.com/bejayjones/juno/internal/identity/application"
	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/bejayjones/juno/pkg/clock"
)

// --- fakes ----------------------------------------------------------------

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
	for email, e := range r.byEmail {
		if e.ID == i.ID && email != i.Email {
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
	var out []*domain.Inspector
	for _, i := range r.byID {
		if i.CompanyID == companyID {
			out = append(out, i)
		}
	}
	return out, nil
}
func (r *fakeInspectorRepo) Delete(_ context.Context, id domain.InspectorID) error {
	if i, ok := r.byID[id]; ok {
		delete(r.byEmail, i.Email)
		delete(r.byID, id)
	}
	return nil
}

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
func (r *fakeClientRepo) FindByCompany(_ context.Context, companyID domain.CompanyID, _ domain.ClientFilter) ([]*domain.Client, error) {
	var out []*domain.Client
	for _, c := range r.byID {
		if c.CompanyID == companyID {
			out = append(out, c)
		}
	}
	return out, nil
}
func (r *fakeClientRepo) Delete(_ context.Context, id domain.ClientID) error {
	delete(r.byID, id)
	return nil
}

type fakeHasher struct{}

func (fakeHasher) Hash(p string) (string, error) { return "hashed:" + p, nil }
func (fakeHasher) Verify(hash, p string) bool    { return hash == "hashed:"+p }

type fakeTokenIssuer struct{}

func (fakeTokenIssuer) Issue(_, _, _ string) (string, time.Time, error) {
	return "test-token", time.Now().Add(24 * time.Hour), nil
}

// fakeTokenVerifier verifies only the hard-coded "test-token" value.
type fakeTokenVerifier struct {
	principal middleware.Principal
}

func (v *fakeTokenVerifier) VerifyToken(token string) (middleware.Principal, error) {
	if token != "test-token" {
		return middleware.Principal{}, http.ErrNoCookie // any error
	}
	return v.principal, nil
}

// --- helpers --------------------------------------------------------------

func newTestServer(t *testing.T) (*rest.Server, *fakeTokenVerifier) {
	t.Helper()
	inspectorRepo := newFakeInspectorRepo()
	companyRepo := newFakeCompanyRepo()
	clientRepo := newFakeClientRepo()
	clk := clock.Fixed(time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC))

	inspectorSvc := identityapp.NewInspectorService(inspectorRepo, companyRepo, fakeHasher{}, fakeTokenIssuer{}, clk)
	companySvc := identityapp.NewCompanyService(companyRepo, clk)
	clientSvc := identityapp.NewClientService(clientRepo, clk)

	verifier := &fakeTokenVerifier{}
	srv := rest.NewServer(nil, nil, inspectorSvc, companySvc, clientSvc, nil, nil, nil, nil, verifier)
	return srv, verifier
}

func postJSON(t *testing.T, srv http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	return rr
}

func getWithToken(t *testing.T, srv http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)
	return rr
}

// --- tests ----------------------------------------------------------------

func TestHandleRegister_Success(t *testing.T) {
	srv, _ := newTestServer(t)

	rr := postJSON(t, srv, "/api/v1/auth/register", map[string]string{
		"first_name":   "Alice",
		"last_name":    "Smith",
		"email":        "alice@example.com",
		"password":     "secret",
		"company_name": "Acme Inspections",
	})

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rr.Code, rr.Body)
	}

	var out identityapp.RegisterOutput
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Inspector.Email != "alice@example.com" {
		t.Errorf("email = %q", out.Inspector.Email)
	}
	if out.Token != "test-token" {
		t.Errorf("token = %q", out.Token)
	}
}

func TestHandleRegister_DuplicateEmail(t *testing.T) {
	srv, _ := newTestServer(t)

	body := map[string]string{
		"first_name":   "Alice",
		"last_name":    "Smith",
		"email":        "dup@example.com",
		"password":     "secret",
		"company_name": "Acme",
	}
	postJSON(t, srv, "/api/v1/auth/register", body)
	rr := postJSON(t, srv, "/api/v1/auth/register", body)

	if rr.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rr.Code)
	}
}

func TestHandleLogin_Success(t *testing.T) {
	srv, _ := newTestServer(t)

	postJSON(t, srv, "/api/v1/auth/register", map[string]string{
		"first_name":   "Alice",
		"last_name":    "Smith",
		"email":        "alice@example.com",
		"password":     "correctpass",
		"company_name": "Acme",
	})

	rr := postJSON(t, srv, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "correctpass",
	})
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rr.Code, rr.Body)
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	srv, _ := newTestServer(t)

	postJSON(t, srv, "/api/v1/auth/register", map[string]string{
		"first_name":   "Alice",
		"last_name":    "Smith",
		"email":        "alice@example.com",
		"password":     "correctpass",
		"company_name": "Acme",
	})

	rr := postJSON(t, srv, "/api/v1/auth/login", map[string]string{
		"email":    "alice@example.com",
		"password": "wrongpass",
	})
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestHandleGetMe_Authenticated(t *testing.T) {
	srv, verifier := newTestServer(t)

	// Register to get an inspector ID.
	rr := postJSON(t, srv, "/api/v1/auth/register", map[string]string{
		"first_name":   "Alice",
		"last_name":    "Smith",
		"email":        "alice@example.com",
		"password":     "pass",
		"company_name": "Acme",
	})
	var regOut identityapp.RegisterOutput
	_ = json.NewDecoder(rr.Body).Decode(&regOut)

	verifier.principal = middleware.Principal{
		InspectorID: regOut.Inspector.ID,
		CompanyID:   regOut.Company.ID,
		Role:        regOut.Inspector.Role,
	}

	meRR := getWithToken(t, srv, "/api/v1/me", "test-token")
	if meRR.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", meRR.Code, meRR.Body)
	}

	var view identityapp.InspectorView
	_ = json.NewDecoder(meRR.Body).Decode(&view)
	if view.Email != "alice@example.com" {
		t.Errorf("email = %q", view.Email)
	}
}

func TestHandleGetMe_Unauthenticated(t *testing.T) {
	srv, _ := newTestServer(t)
	rr := getWithToken(t, srv, "/api/v1/me", "") // no token
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
