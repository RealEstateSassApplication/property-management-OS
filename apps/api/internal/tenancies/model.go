package tenancies

import "time"

type Tenancy struct {
	ID                string    `json:"id"`
	OrganizationID    string    `json:"organizationId"`
	UnitID            string    `json:"unitId"`
	UnitLabel         string    `json:"unitLabel"`
	PropertyName      string    `json:"propertyName"`
	PrimaryTenantID   string    `json:"primaryTenantId"`
	PrimaryTenantName string    `json:"primaryTenantName"`
	OccupantCount     int       `json:"occupantCount"`
	StartDate         string    `json:"startDate"`
	EndDate           string    `json:"endDate,omitempty"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type CreateInput struct {
	UnitID            string   `json:"unitId"`
	PrimaryTenantID   string   `json:"primaryTenantId"`
	OccupantTenantIDs []string `json:"occupantTenantIds"`
	StartDate         string   `json:"startDate"`
	EndDate           string   `json:"endDate"`
	Status            string   `json:"status"`
}

type UpdateInput struct {
	EndDate *string `json:"endDate"`
	Status  *string `json:"status"`
}
