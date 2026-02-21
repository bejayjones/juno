package application

import "github.com/bejayjones/juno/internal/identity/domain"

// LicenseView is the read model for a state license.
type LicenseView struct {
	State  string `json:"state"`
	Number string `json:"number"`
}

// InspectorView is the read model for an Inspector aggregate.
type InspectorView struct {
	ID        string        `json:"id"`
	CompanyID string        `json:"company_id"`
	FirstName string        `json:"first_name"`
	LastName  string        `json:"last_name"`
	Email     string        `json:"email"`
	Role      string        `json:"role"`
	Licenses  []LicenseView `json:"licenses"`
	CreatedAt int64         `json:"created_at"`
	UpdatedAt int64         `json:"updated_at"`
}

func toInspectorView(i *domain.Inspector) InspectorView {
	licenses := make([]LicenseView, len(i.LicenseNumbers))
	for j, l := range i.LicenseNumbers {
		licenses[j] = LicenseView{State: l.State, Number: l.Number}
	}
	return InspectorView{
		ID:        string(i.ID),
		CompanyID: string(i.CompanyID),
		FirstName: i.Name.First,
		LastName:  i.Name.Last,
		Email:     i.Email,
		Role:      string(i.Role),
		Licenses:  licenses,
		CreatedAt: i.CreatedAt.Unix(),
		UpdatedAt: i.UpdatedAt.Unix(),
	}
}

// AddressView is the read model for a postal address.
type AddressView struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
	Country string `json:"country"`
}

// CompanyView is the read model for a Company aggregate.
type CompanyView struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Address         AddressView `json:"address"`
	Phone           string      `json:"phone"`
	Email           string      `json:"email"`
	LogoStoragePath string      `json:"logo_storage_path,omitempty"`
	CreatedAt       int64       `json:"created_at"`
	UpdatedAt       int64       `json:"updated_at"`
}

func toCompanyView(c *domain.Company) CompanyView {
	return CompanyView{
		ID:   string(c.ID),
		Name: c.Name,
		Address: AddressView{
			Street:  c.Address.Street,
			City:    c.Address.City,
			State:   c.Address.State,
			Zip:     c.Address.Zip,
			Country: c.Address.Country,
		},
		Phone:           c.Phone,
		Email:           c.Email,
		LogoStoragePath: c.LogoStoragePath,
		CreatedAt:       c.CreatedAt.Unix(),
		UpdatedAt:       c.UpdatedAt.Unix(),
	}
}

// ClientView is the read model for a Client aggregate.
type ClientView struct {
	ID        string `json:"id"`
	CompanyID string `json:"company_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

func toClientView(c *domain.Client) ClientView {
	return ClientView{
		ID:        string(c.ID),
		CompanyID: string(c.CompanyID),
		FirstName: c.Name.First,
		LastName:  c.Name.Last,
		Email:     c.Email,
		Phone:     c.Phone,
		CreatedAt: c.CreatedAt.Unix(),
		UpdatedAt: c.UpdatedAt.Unix(),
	}
}
