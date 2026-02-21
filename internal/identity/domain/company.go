package domain

import "time"

// Company is the aggregate root for an inspection company. Inspectors belong
// to a company and share its branding on generated reports.
type Company struct {
	ID              CompanyID
	Name            string
	Address         Address
	Phone           string
	Email           string
	LogoStoragePath string // empty until a logo is uploaded
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewCompany(
	id CompanyID,
	name string,
	address Address,
	phone string,
	email string,
	now time.Time,
) *Company {
	return &Company{
		ID:        id,
		Name:      name,
		Address:   address,
		Phone:     phone,
		Email:     email,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Company) UpdateProfile(name string, address Address, phone, email string, now time.Time) {
	c.Name = name
	c.Address = address
	c.Phone = phone
	c.Email = email
	c.UpdatedAt = now
}

func (c *Company) SetLogo(storagePath string, now time.Time) {
	c.LogoStoragePath = storagePath
	c.UpdatedAt = now
}
