package properties

import "time"

type Property struct {
	ID                      string    `json:"id"`
	OrganizationID          string    `json:"organizationId"`
	ReferenceCode           *string   `json:"referenceCode,omitempty"`
	Name                    string    `json:"name"`
	PropertyType            string    `json:"propertyType"`
	AddressLine1            string    `json:"addressLine1"`
	AddressLine2            *string   `json:"addressLine2,omitempty"`
	City                    string    `json:"city"`
	Region                  *string   `json:"region,omitempty"`
	PostalCode              *string   `json:"postalCode,omitempty"`
	CountryCode             string    `json:"countryCode"`
	Status                  string    `json:"status"`
	ExternalAvaraPropertyID *string   `json:"externalAvaraPropertyId,omitempty"`
	CreatedAt               time.Time `json:"createdAt"`
	UpdatedAt               time.Time `json:"updatedAt"`
}

type CreateInput struct {
	ReferenceCode           *string `json:"referenceCode"`
	Name                    string  `json:"name"`
	PropertyType            string  `json:"propertyType"`
	AddressLine1            string  `json:"addressLine1"`
	AddressLine2            *string `json:"addressLine2"`
	City                    string  `json:"city"`
	Region                  *string `json:"region"`
	PostalCode              *string `json:"postalCode"`
	CountryCode             string  `json:"countryCode"`
	ExternalAvaraPropertyID *string `json:"externalAvaraPropertyId"`
}

type UpdateInput struct {
	ReferenceCode *string `json:"referenceCode"`
	Name          *string `json:"name"`
	PropertyType  *string `json:"propertyType"`
	AddressLine1  *string `json:"addressLine1"`
	AddressLine2  *string `json:"addressLine2"`
	City          *string `json:"city"`
	Region        *string `json:"region"`
	PostalCode    *string `json:"postalCode"`
	CountryCode   *string `json:"countryCode"`
	Status        *string `json:"status"`
}
