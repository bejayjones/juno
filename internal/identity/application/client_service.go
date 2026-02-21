package application

import (
	"context"
	"fmt"

	"github.com/bejayjones/juno/internal/identity/domain"
	"github.com/bejayjones/juno/pkg/clock"
	"github.com/bejayjones/juno/pkg/id"
)

// ClientService manages the client roster for a company.
type ClientService struct {
	clients domain.ClientRepository
	clock   clock.Clock
}

func NewClientService(clients domain.ClientRepository, clk clock.Clock) *ClientService {
	return &ClientService{clients: clients, clock: clk}
}

// CreateClientInput contains the fields needed to add a client to a company.
type CreateClientInput struct {
	CompanyID string
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

// Create adds a new client to a company's roster.
func (s *ClientService) Create(ctx context.Context, in CreateClientInput) (ClientView, error) {
	client := domain.NewClient(
		domain.ClientID(id.New()),
		domain.CompanyID(in.CompanyID),
		domain.PersonName{First: in.FirstName, Last: in.LastName},
		in.Email,
		in.Phone,
		s.clock.Now(),
	)
	if err := s.clients.Save(ctx, client); err != nil {
		return ClientView{}, fmt.Errorf("save client: %w", err)
	}
	return toClientView(client), nil
}

// GetByID returns a single client.
func (s *ClientService) GetByID(ctx context.Context, clientID domain.ClientID) (ClientView, error) {
	client, err := s.clients.FindByID(ctx, clientID)
	if err != nil {
		return ClientView{}, err
	}
	return toClientView(client), nil
}

// List returns clients for a company, supporting search and pagination via filter.
func (s *ClientService) List(ctx context.Context, companyID domain.CompanyID, filter domain.ClientFilter) ([]ClientView, error) {
	clients, err := s.clients.FindByCompany(ctx, companyID, filter)
	if err != nil {
		return nil, fmt.Errorf("list clients: %w", err)
	}
	views := make([]ClientView, len(clients))
	for i, c := range clients {
		views[i] = toClientView(c)
	}
	return views, nil
}

// UpdateClientInput contains the mutable contact fields for a client.
type UpdateClientInput struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

// Update changes a client's contact details.
func (s *ClientService) Update(ctx context.Context, clientID domain.ClientID, in UpdateClientInput) (ClientView, error) {
	client, err := s.clients.FindByID(ctx, clientID)
	if err != nil {
		return ClientView{}, err
	}

	client.UpdateContact(
		domain.PersonName{First: in.FirstName, Last: in.LastName},
		in.Email, in.Phone,
		s.clock.Now(),
	)
	if err := s.clients.Save(ctx, client); err != nil {
		return ClientView{}, fmt.Errorf("save client: %w", err)
	}
	return toClientView(client), nil
}

// Delete removes a client from the roster.
func (s *ClientService) Delete(ctx context.Context, clientID domain.ClientID) error {
	return s.clients.Delete(ctx, clientID)
}
