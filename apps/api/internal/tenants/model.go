package tenants

import "time"

type Tenant struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	LegalName      string    `json:"legalName"`
	Email          string    `json:"email,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type CreateInput struct {
	LegalName string `json:"legalName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
}

type UpdateInput struct {
	LegalName *string `json:"legalName"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Status    *string `json:"status"`
}
