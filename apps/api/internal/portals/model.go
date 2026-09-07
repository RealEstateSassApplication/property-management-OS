package portals

import (
	"errors"
	"time"
)

type OwnerProfile struct {
	ID        string `json:"id"`
	LegalName string `json:"legalName"`
	OwnerType string `json:"ownerType"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
}

type OwnerProperty struct {
	OwnerID               string           `json:"ownerId"`
	PropertyID            string           `json:"propertyId"`
	ReferenceCode         string           `json:"referenceCode"`
	Name                  string           `json:"name"`
	City                  string           `json:"city,omitempty"`
	OwnershipBPS          int              `json:"ownershipBps"`
	UnitCount             int              `json:"unitCount"`
	OccupiedUnits         int              `json:"occupiedUnits"`
	OpenMaintenance       int              `json:"openMaintenance"`
	OutstandingByCurrency map[string]int64 `json:"outstandingByCurrencyMinor"`
}

type OwnerSummary struct {
	Owners     []OwnerProfile  `json:"owners"`
	Properties []OwnerProperty `json:"properties"`
}

type TenantProfile struct {
	ID        string `json:"id"`
	LegalName string `json:"legalName"`
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Status    string `json:"status"`
}

type TenantOccupancy struct {
	TenantID           string `json:"tenantId"`
	TenancyID          string `json:"tenancyId"`
	OccupantRole       string `json:"occupantRole"`
	TenancyStatus      string `json:"tenancyStatus"`
	StartDate          string `json:"startDate"`
	EndDate            string `json:"endDate,omitempty"`
	PropertyID         string `json:"propertyId"`
	PropertyName       string `json:"propertyName"`
	UnitID             string `json:"unitId"`
	UnitLabel          string `json:"unitLabel"`
	LeaseID            string `json:"leaseId,omitempty"`
	LeaseReference     string `json:"leaseReference,omitempty"`
	LeaseStatus        string `json:"leaseStatus,omitempty"`
	LeaseStartDate     string `json:"leaseStartDate,omitempty"`
	LeaseEndDate       string `json:"leaseEndDate,omitempty"`
	RentAmountMinor    int64  `json:"rentAmountMinor,omitempty"`
	DepositAmountMinor int64  `json:"depositAmountMinor,omitempty"`
	Currency           string `json:"currency,omitempty"`
	DueDay             int    `json:"dueDay,omitempty"`
}

type TenantRentItem struct {
	ObligationID   string `json:"obligationId"`
	TenantID       string `json:"tenantId"`
	TenancyID      string `json:"tenancyId"`
	LeaseID        string `json:"leaseId"`
	Period         string `json:"period"`
	DueDate        string `json:"dueDate"`
	AmountMinor    int64  `json:"amountMinor"`
	AllocatedMinor int64  `json:"allocatedMinor"`
	BalanceMinor   int64  `json:"balanceMinor"`
	Currency       string `json:"currency"`
	State          string `json:"state"`
}

type TenantMaintenanceItem struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	PropertyName string    `json:"propertyName"`
	UnitLabel    string    `json:"unitLabel,omitempty"`
	Title        string    `json:"title"`
	Category     string    `json:"category"`
	Priority     string    `json:"priority"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type TenantSummary struct {
	Tenants     []TenantProfile         `json:"tenants"`
	Occupancies []TenantOccupancy       `json:"occupancies"`
	Rent        []TenantRentItem        `json:"rent"`
	Maintenance []TenantMaintenanceItem `json:"maintenance"`
}

type CreateTenantMaintenanceInput struct {
	TenancyID   string `json:"tenancyId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

type TenantMaintenanceContext struct {
	TenantID   string
	PropertyID string
	UnitID     string
}

type OwnerPropertyBase struct {
	OwnerID         string
	PropertyID      string
	ReferenceCode   string
	Name            string
	City            string
	OwnershipBPS    int
	UnitCount       int
	OccupiedUnits   int
	OpenMaintenance int
}

type OwnerReceivableRow struct {
	PropertyID        string
	Currency          string
	OutstandingMinor int64
}

var (
	ErrOwnerLinkNotFound    = errors.New("no owner record is linked to this user")
	ErrTenantLinkNotFound   = errors.New("no tenant record is linked to this user")
	ErrTenancyNotAccessible = errors.New("tenancy is not an active occupancy linked to this user")
)
