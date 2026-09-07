package orgadmin

import (
	"errors"
	"time"
)

type OrganizationSettings struct {
	OrganizationID  string    `json:"organizationId"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Status          string    `json:"status"`
	Timezone        string    `json:"timezone"`
	DefaultCurrency string    `json:"defaultCurrency"`
	CountryCode     string    `json:"countryCode"`
	BillingEmail    string    `json:"billingEmail,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Member struct {
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	UserStatus  string    `json:"userStatus"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"createdAt"`
}

type PortalLink struct {
	Kind         string `json:"kind"`
	UserID       string `json:"userId"`
	Email        string `json:"email"`
	DisplayName  string `json:"displayName"`
	ResourceID   string `json:"resourceId"`
	ResourceName string `json:"resourceName"`
}

type UpdateOrganizationInput struct {
	Name            string `json:"name"`
	Timezone        string `json:"timezone"`
	DefaultCurrency string `json:"defaultCurrency"`
	CountryCode     string `json:"countryCode"`
	BillingEmail    string `json:"billingEmail"`
}

type CreateMemberInput struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type UpdateMemberRoleInput struct {
	Role string `json:"role"`
}

type CreateOwnerPortalLinkInput struct {
	UserID  string `json:"userId"`
	OwnerID string `json:"ownerId"`
}

type CreateTenantPortalLinkInput struct {
	UserID   string `json:"userId"`
	TenantID string `json:"tenantId"`
}

var (
	ErrMemberNotFound    = errors.New("organization member not found")
	ErrLastAdmin         = errors.New("the final organization admin cannot be removed or demoted")
	ErrInvalidRole       = errors.New("unsupported organization role")
	ErrPortalLinkInvalid = errors.New("portal link requires a matching owner/tenant role and resource in this organization")
)
