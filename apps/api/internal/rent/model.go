package rent

import (
	"errors"
	"time"
)

type Obligation struct {
	ID                string    `json:"id"`
	OrganizationID    string    `json:"organizationId"`
	LeaseID           string    `json:"leaseId"`
	LeaseReference    string    `json:"leaseReference"`
	PropertyName      string    `json:"propertyName"`
	UnitLabel         string    `json:"unitLabel"`
	PrimaryTenantName string    `json:"primaryTenantName"`
	Period            string    `json:"period"`
	DueDate           string    `json:"dueDate"`
	AmountMinor       int64     `json:"amountMinor"`
	AllocatedMinor    int64     `json:"allocatedMinor"`
	BalanceMinor      int64     `json:"balanceMinor"`
	Currency          string    `json:"currency"`
	State             string    `json:"state"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Payment struct {
	ID               string    `json:"id"`
	OrganizationID   string    `json:"organizationId"`
	TenantID         string    `json:"tenantId"`
	TenantName       string    `json:"tenantName"`
	AmountMinor      int64     `json:"amountMinor"`
	AllocatedMinor   int64     `json:"allocatedMinor"`
	UnallocatedMinor int64     `json:"unallocatedMinor"`
	Currency         string    `json:"currency"`
	ReceivedAt       string    `json:"receivedAt"`
	Method           string    `json:"method"`
	ReferenceCode    string    `json:"referenceCode,omitempty"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"createdAt"`
}

type Allocation struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	PaymentID      string    `json:"paymentId"`
	ObligationID   string    `json:"obligationId"`
	AmountMinor    int64     `json:"amountMinor"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateObligationInput struct {
	LeaseID string `json:"leaseId"`
	Period  string `json:"period"`
}

type CreatePaymentInput struct {
	TenantID      string `json:"tenantId"`
	AmountMinor   int64  `json:"amountMinor"`
	Currency      string `json:"currency"`
	ReceivedAt    string `json:"receivedAt"`
	Method        string `json:"method"`
	ReferenceCode string `json:"referenceCode"`
}

type CreateAllocationInput struct {
	PaymentID    string `json:"paymentId"`
	ObligationID string `json:"obligationId"`
	AmountMinor  int64  `json:"amountMinor"`
}

var (
	ErrLeaseNotFound               = errors.New("lease not found")
	ErrLeaseNotActive              = errors.New("lease must be active")
	ErrPeriodOutsideLease          = errors.New("period falls outside the lease term")
	ErrObligationExists            = errors.New("rent obligation already exists for this lease and period")
	ErrTenantNotFound              = errors.New("tenant not found")
	ErrPaymentNotFound             = errors.New("payment not found")
	ErrObligationNotFound          = errors.New("rent obligation not found")
	ErrPaymentNotPosted            = errors.New("payment is not posted")
	ErrObligationVoided            = errors.New("rent obligation is void")
	ErrCurrencyMismatch            = errors.New("payment and obligation currencies differ")
	ErrAllocationExceedsPayment    = errors.New("allocation exceeds unallocated payment balance")
	ErrAllocationExceedsObligation = errors.New("allocation exceeds outstanding obligation balance")
	ErrTenantMismatch              = errors.New("payment tenant does not belong to the obligation tenancy")
)
