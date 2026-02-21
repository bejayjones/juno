package domain

import "time"

type ClientID string

// Client is a person who commissions an inspection. Clients belong to a company's
// roster and may be associated with multiple inspections over time.
type Client struct {
	ID        ClientID
	CompanyID CompanyID
	Name      PersonName
	Email     string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewClient(
	id ClientID,
	companyID CompanyID,
	name PersonName,
	email string,
	phone string,
	now time.Time,
) *Client {
	return &Client{
		ID:        id,
		CompanyID: companyID,
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (c *Client) UpdateContact(name PersonName, email, phone string, now time.Time) {
	c.Name = name
	c.Email = email
	c.Phone = phone
	c.UpdatedAt = now
}
