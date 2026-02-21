package domain

import (
	"fmt"
	"strings"
)

// PersonName is the full name of an inspector or client.
type PersonName struct {
	First string
	Last  string
}

func (n PersonName) Full() string {
	return strings.TrimSpace(n.First + " " + n.Last)
}

// Address is a postal address.
type Address struct {
	Street  string
	City    string
	State   string
	Zip     string
	Country string
}

func (a Address) String() string {
	return fmt.Sprintf("%s, %s, %s %s", a.Street, a.City, a.State, a.Zip)
}

// LicenseNumber is a state-scoped inspector license.
type LicenseNumber struct {
	State  string
	Number string
}

// InspectorRole controls access within a company.
type InspectorRole string

const (
	RoleOwner  InspectorRole = "owner"
	RoleMember InspectorRole = "member"
)

func (r InspectorRole) Valid() bool {
	return r == RoleOwner || r == RoleMember
}
