package maintenance

import (
	"errors"
	"time"
)

type Vendor struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Trade          string    `json:"trade"`
	Email          string    `json:"email,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type Request struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	PropertyID     string    `json:"propertyId"`
	PropertyName   string    `json:"propertyName"`
	UnitID         string    `json:"unitId,omitempty"`
	UnitLabel      string    `json:"unitLabel,omitempty"`
	TenantID       string    `json:"tenantId,omitempty"`
	TenantName     string    `json:"tenantName,omitempty"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Category       string    `json:"category"`
	Priority       string    `json:"priority"`
	Status         string    `json:"status"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type WorkOrder struct {
	ID                   string     `json:"id"`
	OrganizationID       string     `json:"organizationId"`
	MaintenanceRequestID string     `json:"maintenanceRequestId"`
	RequestTitle         string     `json:"requestTitle"`
	PropertyName         string     `json:"propertyName"`
	UnitLabel            string     `json:"unitLabel,omitempty"`
	VendorID             string     `json:"vendorId,omitempty"`
	VendorName           string     `json:"vendorName,omitempty"`
	AssignedUserID       string     `json:"assignedUserId,omitempty"`
	Summary              string     `json:"summary"`
	Status               string     `json:"status"`
	ScheduledFor         *time.Time `json:"scheduledFor,omitempty"`
	StartedAt            *time.Time `json:"startedAt,omitempty"`
	CompletedAt          *time.Time `json:"completedAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type Quote struct {
	ID               string     `json:"id"`
	OrganizationID   string     `json:"organizationId"`
	WorkOrderID      string     `json:"workOrderId"`
	WorkOrderSummary string     `json:"workOrderSummary"`
	VendorID         string     `json:"vendorId"`
	VendorName       string     `json:"vendorName"`
	AmountMinor      int64      `json:"amountMinor"`
	Currency         string     `json:"currency"`
	ScopeSummary     string     `json:"scopeSummary"`
	Status           string     `json:"status"`
	SubmittedAt      time.Time  `json:"submittedAt"`
	ReviewedAt       *time.Time `json:"reviewedAt,omitempty"`
	ReviewedByUserID string     `json:"reviewedByUserId,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type Evidence struct {
	ID                string    `json:"id"`
	OrganizationID    string    `json:"organizationId"`
	WorkOrderID       string    `json:"workOrderId"`
	EvidenceType      string    `json:"evidenceType"`
	Note              string    `json:"note,omitempty"`
	StorageKey        string    `json:"storageKey,omitempty"`
	SubmittedByUserID string    `json:"submittedByUserId,omitempty"`
	CreatedAt         time.Time `json:"createdAt"`
}

type CreateVendorInput struct {
	Name   string `json:"name"`
	Trade  string `json:"trade"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Status string `json:"status"`
}

type CreateRequestInput struct {
	PropertyID  string `json:"propertyId"`
	UnitID      string `json:"unitId"`
	TenantID    string `json:"tenantId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Priority    string `json:"priority"`
}

type UpdateRequestStatusInput struct {
	Status string `json:"status"`
}

type CreateWorkOrderInput struct {
	MaintenanceRequestID string `json:"maintenanceRequestId"`
	VendorID             string `json:"vendorId"`
	AssignedUserID       string `json:"assignedUserId"`
	Summary              string `json:"summary"`
	ScheduledFor         string `json:"scheduledFor"`
}

type UpdateWorkOrderStatusInput struct {
	Status string `json:"status"`
}

type CreateQuoteInput struct {
	WorkOrderID  string `json:"workOrderId"`
	VendorID     string `json:"vendorId"`
	AmountMinor  int64  `json:"amountMinor"`
	Currency     string `json:"currency"`
	ScopeSummary string `json:"scopeSummary"`
}

type DecideQuoteInput struct {
	Decision string `json:"decision"`
}

type CreateEvidenceInput struct {
	WorkOrderID  string `json:"workOrderId"`
	EvidenceType string `json:"evidenceType"`
	Note         string `json:"note"`
	StorageKey   string `json:"storageKey"`
}

var (
	ErrPropertyNotFound          = errors.New("property not found")
	ErrUnitNotFound              = errors.New("unit not found for property")
	ErrTenantNotFound            = errors.New("tenant not found")
	ErrTenantNotOccupant         = errors.New("tenant is not an active occupant of the selected unit")
	ErrVendorNotFound            = errors.New("vendor not found")
	ErrVendorInactive            = errors.New("vendor is inactive")
	ErrRequestNotFound           = errors.New("maintenance request not found")
	ErrRequestClosed             = errors.New("maintenance request is already closed")
	ErrInvalidRequestTransition  = errors.New("invalid maintenance request status transition")
	ErrWorkOrderNotFound         = errors.New("work order not found")
	ErrInvalidWorkOrderTransition = errors.New("invalid work order status transition")
	ErrCompletionEvidenceRequired = errors.New("completion evidence is required before completing a work order")
	ErrQuoteNotFound             = errors.New("maintenance quote not found")
	ErrQuoteNotSubmitted         = errors.New("only submitted quotes can be decided")
	ErrApprovedQuoteExists       = errors.New("an approved quote already exists for this work order")
	ErrWorkOrderClosed           = errors.New("work order is already completed or cancelled")
)
