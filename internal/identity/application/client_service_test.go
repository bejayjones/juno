package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/bejayjones/juno/internal/identity/application"
	"github.com/bejayjones/juno/internal/identity/domain"
)

func newTestClientService() *application.ClientService {
	return application.NewClientService(newFakeClientRepo(), fixedClock{t: testNow})
}

func TestClientService_Create(t *testing.T) {
	svc := newTestClientService()

	view, err := svc.Create(context.Background(), application.CreateClientInput{
		CompanyID: "co-1",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Phone:     "555-1234",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if view.FirstName != "Jane" || view.LastName != "Doe" {
		t.Errorf("name = %q %q", view.FirstName, view.LastName)
	}
	if view.CompanyID != "co-1" {
		t.Errorf("company_id = %q", view.CompanyID)
	}
	if view.ID == "" {
		t.Error("expected non-empty id")
	}
}

func TestClientService_GetByID_NotFound(t *testing.T) {
	svc := newTestClientService()
	_, err := svc.GetByID(context.Background(), domain.ClientID("nonexistent"))
	if !errors.Is(err, domain.ErrClientNotFound) {
		t.Errorf("want ErrClientNotFound, got %v", err)
	}
}

func TestClientService_List(t *testing.T) {
	svc := newTestClientService()
	for i := range 3 {
		_, _ = svc.Create(context.Background(), application.CreateClientInput{
			CompanyID: "co-1",
			FirstName: "Client",
			LastName:  string(rune('A' + i)),
		})
	}
	// Different company — should not appear.
	_, _ = svc.Create(context.Background(), application.CreateClientInput{
		CompanyID: "co-2",
		FirstName: "Other",
		LastName:  "Co",
	})

	views, err := svc.List(context.Background(), domain.CompanyID("co-1"), domain.ClientFilter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(views) != 3 {
		t.Errorf("want 3 clients for co-1, got %d", len(views))
	}
}

func TestClientService_Update(t *testing.T) {
	svc := newTestClientService()
	created, _ := svc.Create(context.Background(), application.CreateClientInput{
		CompanyID: "co-1",
		FirstName: "Jane",
		LastName:  "Doe",
	})

	updated, err := svc.Update(context.Background(), domain.ClientID(created.ID), application.UpdateClientInput{
		FirstName: "Janet",
		LastName:  "Doe",
		Email:     "janet@example.com",
		Phone:     "555-9999",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.FirstName != "Janet" {
		t.Errorf("first_name = %q", updated.FirstName)
	}
	if updated.Email != "janet@example.com" {
		t.Errorf("email = %q", updated.Email)
	}
}

func TestClientService_Delete(t *testing.T) {
	svc := newTestClientService()
	created, _ := svc.Create(context.Background(), application.CreateClientInput{
		CompanyID: "co-1",
		FirstName: "Jane",
		LastName:  "Doe",
	})

	if err := svc.Delete(context.Background(), domain.ClientID(created.ID)); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := svc.GetByID(context.Background(), domain.ClientID(created.ID))
	if !errors.Is(err, domain.ErrClientNotFound) {
		t.Errorf("want ErrClientNotFound after delete, got %v", err)
	}
}
