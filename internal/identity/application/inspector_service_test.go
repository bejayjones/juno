package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bejayjones/juno/internal/identity/application"
	"github.com/bejayjones/juno/internal/identity/domain"
)

func newTestInspectorService() *application.InspectorService {
	return application.NewInspectorService(
		newFakeInspectorRepo(),
		newFakeCompanyRepo(),
		fakeHasher{},
		fakeTokenIssuer{},
		fixedClock{t: testNow},
	)
}

func TestInspectorService_Register_CreatesCompanyAndOwner(t *testing.T) {
	svc := newTestInspectorService()

	out, err := svc.Register(context.Background(), application.RegisterInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		Password:    "secret",
		CompanyName: "Acme Inspections",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if out.Inspector.Email != "alice@example.com" {
		t.Errorf("email = %q, want %q", out.Inspector.Email, "alice@example.com")
	}
	if out.Inspector.Role != string(domain.RoleOwner) {
		t.Errorf("role = %q, want owner", out.Inspector.Role)
	}
	if out.Company.Name != "Acme Inspections" {
		t.Errorf("company name = %q, want %q", out.Company.Name, "Acme Inspections")
	}
	if out.Token != "test-token" {
		t.Errorf("token = %q, want test-token", out.Token)
	}
	if out.Inspector.CompanyID != out.Company.ID {
		t.Errorf("inspector company_id %q != company id %q", out.Inspector.CompanyID, out.Company.ID)
	}
}

func TestInspectorService_Register_JoinsExistingCompanyAsMember(t *testing.T) {
	inspectorRepo := newFakeInspectorRepo()
	companyRepo := newFakeCompanyRepo()
	svc := application.NewInspectorService(inspectorRepo, companyRepo, fakeHasher{}, fakeTokenIssuer{}, fixedClock{t: testNow})

	// Seed a company.
	company := domain.NewCompany(domain.CompanyID("co-1"), "Acme", domain.Address{}, "", "", testNow)
	_ = companyRepo.Save(context.Background(), company)

	out, err := svc.Register(context.Background(), application.RegisterInput{
		FirstName: "Bob",
		LastName:  "Jones",
		Email:     "bob@example.com",
		Password:  "secret",
		CompanyID: "co-1",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if out.Inspector.Role != string(domain.RoleMember) {
		t.Errorf("role = %q, want member", out.Inspector.Role)
	}
}

func TestInspectorService_Register_DuplicateEmailReturnsErrEmailTaken(t *testing.T) {
	svc := newTestInspectorService()

	in := application.RegisterInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "dup@example.com",
		Password:    "secret",
		CompanyName: "First Co",
	}
	if _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatalf("first register: %v", err)
	}

	in.CompanyName = "Second Co"
	_, err := svc.Register(context.Background(), in)
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("want ErrEmailTaken, got %v", err)
	}
}

func TestInspectorService_Register_MissingCompanyNameReturnsError(t *testing.T) {
	svc := newTestInspectorService()
	_, err := svc.Register(context.Background(), application.RegisterInput{
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
		Password:  "secret",
		// CompanyID and CompanyName both empty
	})
	if err == nil {
		t.Error("expected error when company_name is missing")
	}
}

func TestInspectorService_Login_Success(t *testing.T) {
	svc := newTestInspectorService()
	_, err := svc.Register(context.Background(), application.RegisterInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		Password:    "correctpass",
		CompanyName: "Acme",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	out, err := svc.Login(context.Background(), "alice@example.com", "correctpass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if out.Inspector.Email != "alice@example.com" {
		t.Errorf("email = %q", out.Inspector.Email)
	}
	if out.Token != "test-token" {
		t.Errorf("token = %q", out.Token)
	}
}

func TestInspectorService_Login_WrongPasswordReturnsNotFound(t *testing.T) {
	svc := newTestInspectorService()
	_, _ = svc.Register(context.Background(), application.RegisterInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		Password:    "correctpass",
		CompanyName: "Acme",
	})

	_, err := svc.Login(context.Background(), "alice@example.com", "wrongpass")
	if !errors.Is(err, domain.ErrInspectorNotFound) {
		t.Errorf("want ErrInspectorNotFound, got %v", err)
	}
}

func TestInspectorService_UpdateProfile(t *testing.T) {
	svc := newTestInspectorService()
	out, _ := svc.Register(context.Background(), application.RegisterInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		Password:    "pass",
		CompanyName: "Acme",
	})

	updated, err := svc.UpdateProfile(context.Background(), domain.InspectorID(out.Inspector.ID), application.UpdateProfileInput{
		FirstName: "Alicia",
		LastName:  "Jones",
		Email:     "alicia@example.com",
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.FirstName != "Alicia" {
		t.Errorf("first_name = %q, want Alicia", updated.FirstName)
	}
	if updated.Email != "alicia@example.com" {
		t.Errorf("email = %q", updated.Email)
	}
}

func TestInspectorService_SetLicense(t *testing.T) {
	svc := newTestInspectorService()
	out, _ := svc.Register(context.Background(), application.RegisterInput{
		FirstName:   "Alice",
		LastName:    "Smith",
		Email:       "alice@example.com",
		Password:    "pass",
		CompanyName: "Acme",
	})

	updated, err := svc.SetLicense(context.Background(), domain.InspectorID(out.Inspector.ID), application.SetLicenseInput{
		State:  "TX",
		Number: "TX-12345",
	})
	if err != nil {
		t.Fatalf("SetLicense: %v", err)
	}
	if len(updated.Licenses) != 1 {
		t.Fatalf("want 1 license, got %d", len(updated.Licenses))
	}
	if updated.Licenses[0].State != "TX" || updated.Licenses[0].Number != "TX-12345" {
		t.Errorf("unexpected license: %+v", updated.Licenses[0])
	}
}
