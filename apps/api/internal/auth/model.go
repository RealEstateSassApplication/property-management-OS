package auth

import "errors"

type Permission string

const (
	ViewPortfolio   Permission = "portfolio:view"
	ManagePortfolio Permission = "portfolio:manage"
	ViewPeople      Permission = "people:view"
	ManagePeople    Permission = "people:manage"
	ViewLeases      Permission = "leases:view"
	ManageLeases    Permission = "leases:manage"
	ViewOwners      Permission = "owners:view"
	ManageOwners    Permission = "owners:manage"
	ViewRent        Permission = "rent:view"
	ManageRent      Permission = "rent:manage"
)

type Membership struct {
	OrganizationID string `json:"organizationId"`
	UserID         string `json:"userId"`
	Role           string `json:"role"`
}

var (
	ErrMembershipNotFound = errors.New("organization membership not found")
	ErrForbidden          = errors.New("permission denied")
)
