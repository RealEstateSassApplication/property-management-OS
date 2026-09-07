package inspections

import (
	"errors"
	"time"
)

type Inspection struct {
	ID                   string    `json:"id"`
	OrganizationID       string    `json:"organizationId"`
	TenancyID            string    `json:"tenancyId"`
	PropertyID           string    `json:"propertyId"`
	PropertyName         string    `json:"propertyName"`
	UnitID               string    `json:"unitId"`
	UnitLabel            string    `json:"unitLabel"`
	PrimaryTenantName    string    `json:"primaryTenantName"`
	InspectionType       string    `json:"inspectionType"`
	Status               string    `json:"status"`
	ScheduledFor         string    `json:"scheduledFor,omitempty"`
	Summary              string    `json:"summary,omitempty"`
	CompletedAt          string    `json:"completedAt,omitempty"`
	AcknowledgedAt       string    `json:"acknowledgedAt,omitempty"`
	CompletedByUserID    string    `json:"completedByUserId,omitempty"`
	AcknowledgedByUserID string    `json:"acknowledgedByUserId,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type Item struct {
	ID                 string    `json:"id"`
	OrganizationID     string    `json:"organizationId"`
	InspectionID       string    `json:"inspectionId"`
	Area               string    `json:"area"`
	ItemName           string    `json:"itemName"`
	Condition          string    `json:"condition"`
	Notes              string    `json:"notes,omitempty"`
	EvidenceDocumentID string    `json:"evidenceDocumentId,omitempty"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type CreateInspectionInput struct {
	TenancyID      string `json:"tenancyId"`
	InspectionType string `json:"inspectionType"`
	ScheduledFor   string `json:"scheduledFor"`
	Summary        string `json:"summary"`
}

type CreateItemInput struct {
	InspectionID       string `json:"inspectionId"`
	Area               string `json:"area"`
	ItemName           string `json:"itemName"`
	Condition          string `json:"condition"`
	Notes              string `json:"notes"`
	EvidenceDocumentID string `json:"evidenceDocumentId"`
}

type CompleteInspectionInput struct {
	InspectionID string `json:"inspectionId"`
	Summary      string `json:"summary"`
}

type AcknowledgeInspectionInput struct {
	InspectionID string `json:"inspectionId"`
}

var (
	ErrTenancyNotFound          = errors.New("tenancy not found")
	ErrInspectionNotFound       = errors.New("inspection not found")
	ErrInspectionClosed         = errors.New("inspection is already closed")
	ErrInspectionNotCompleted   = errors.New("inspection must be completed first")
	ErrInspectionHasNoItems     = errors.New("inspection requires at least one condition item")
	ErrDocumentNotFound         = errors.New("evidence document not found")
	ErrDocumentNotAvailable     = errors.New("evidence document is not available")
	ErrInvalidInspectionType    = errors.New("unsupported inspection type")
	ErrInvalidInspectionStatus  = errors.New("unsupported inspection status")
	ErrInvalidInspectionItem    = errors.New("invalid inspection item")
)
