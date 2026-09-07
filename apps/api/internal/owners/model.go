package owners

import (
	"errors"
	"time"
)

type Owner struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	LegalName      string    `json:"legalName"`
	OwnerType      string    `json:"ownerType"`
	Email          string    `json:"email,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type OwnershipInterest struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organizationId"`
	PropertyID     string    `json:"propertyId"`
	PropertyName   string    `json:"propertyName"`
	OwnerID        string    `json:"ownerId"`
	OwnerName      string    `json:"ownerName"`
	OwnershipBps   int       `json:"ownershipBps"`
	EffectiveFrom  string    `json:"effectiveFrom"`
	EffectiveTo    string    `json:"effectiveTo,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

type CreateOwnerInput struct {
	LegalName string `json:"legalName"`
	OwnerType string `json:"ownerType"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
}

type CreateInterestInput struct {
	OwnerID       string `json:"ownerId"`
	PropertyID    string `json:"propertyId"`
	OwnershipBps  int    `json:"ownershipBps"`
	EffectiveFrom string `json:"effectiveFrom"`
}

var (
	ErrNotFound              = errors.New("owner not found")
	ErrOwnerNotFound         = errors.New("owner not found")
	ErrPropertyNotFound      = errors.New("property not found")
	ErrOwnershipExceeded     = errors.New("current ownership would exceed 100 percent")
	ErrCurrentInterestExists = errors.New("current ownership interest already exists")
)
