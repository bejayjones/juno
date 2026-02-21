package domain

import "time"

type InspectorID string
type CompanyID string

// Inspector is the aggregate root for an individual licensed home inspector.
type Inspector struct {
	ID             InspectorID
	CompanyID      CompanyID
	Name           PersonName
	Email          string
	PasswordHash   string
	LicenseNumbers []LicenseNumber
	Role           InspectorRole
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewInspector(
	id InspectorID,
	companyID CompanyID,
	name PersonName,
	email string,
	passwordHash string,
	role InspectorRole,
	now time.Time,
) (*Inspector, error) {
	if !role.Valid() {
		return nil, ErrInvalidRole
	}
	return &Inspector{
		ID:           id,
		CompanyID:    companyID,
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (i *Inspector) AddLicense(license LicenseNumber, now time.Time) {
	for _, l := range i.LicenseNumbers {
		if l.State == license.State {
			return // already have a license for this state; update instead
		}
	}
	i.LicenseNumbers = append(i.LicenseNumbers, license)
	i.UpdatedAt = now
}

func (i *Inspector) SetLicense(license LicenseNumber, now time.Time) {
	for j := range i.LicenseNumbers {
		if i.LicenseNumbers[j].State == license.State {
			i.LicenseNumbers[j] = license
			i.UpdatedAt = now
			return
		}
	}
	i.AddLicense(license, now)
}

func (i *Inspector) UpdateProfile(name PersonName, email string, now time.Time) {
	i.Name = name
	i.Email = email
	i.UpdatedAt = now
}

func (i *Inspector) SetPasswordHash(hash string, now time.Time) {
	i.PasswordHash = hash
	i.UpdatedAt = now
}
