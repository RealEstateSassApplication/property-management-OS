package leases

import "time"

type Lease struct {
	ID                 string    `json:"id"`
	OrganizationID     string    `json:"organizationId"`
	TenancyID          string    `json:"tenancyId"`
	ReferenceCode      string    `json:"referenceCode"`
	PropertyName       string    `json:"propertyName"`
	UnitLabel          string    `json:"unitLabel"`
	PrimaryTenantName  string    `json:"primaryTenantName"`
	StartDate          string    `json:"startDate"`
	EndDate            string    `json:"endDate"`
	RentAmountMinor    int64     `json:"rentAmountMinor"`
	DepositAmountMinor int64     `json:"depositAmountMinor"`
	Currency           string    `json:"currency"`
	DueDay             int       `json:"dueDay"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type CreateInput struct {
	TenancyID          string `json:"tenancyId"`
	ReferenceCode      string `json:"referenceCode"`
	StartDate          string `json:"startDate"`
	EndDate            string `json:"endDate"`
	RentAmountMinor    int64  `json:"rentAmountMinor"`
	DepositAmountMinor int64  `json:"depositAmountMinor"`
	Currency           string `json:"currency"`
	DueDay             int    `json:"dueDay"`
	Status             string `json:"status"`
}

type UpdateInput struct {
	Status *string `json:"status"`
}
