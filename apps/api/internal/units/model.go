package units

import "time"

type Unit struct {
	ID              string    `json:"id"`
	OrganizationID  string    `json:"organizationId"`
	PropertyID      string    `json:"propertyId"`
	ReferenceCode   string    `json:"referenceCode"`
	Label           string    `json:"label"`
	Bedrooms        *float64  `json:"bedrooms,omitempty"`
	Bathrooms       *float64  `json:"bathrooms,omitempty"`
	FloorArea       *float64  `json:"floorArea,omitempty"`
	FloorAreaUnit   *string   `json:"floorAreaUnit,omitempty"`
	OccupancyStatus string    `json:"occupancyStatus"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CreateInput struct {
	ReferenceCode string   `json:"referenceCode"`
	Label         string   `json:"label"`
	Bedrooms      *float64 `json:"bedrooms"`
	Bathrooms     *float64 `json:"bathrooms"`
	FloorArea     *float64 `json:"floorArea"`
	FloorAreaUnit *string  `json:"floorAreaUnit"`
}

type UpdateInput struct {
	ReferenceCode   *string  `json:"referenceCode"`
	Label           *string  `json:"label"`
	Bedrooms        *float64 `json:"bedrooms"`
	Bathrooms       *float64 `json:"bathrooms"`
	FloorArea       *float64 `json:"floorArea"`
	FloorAreaUnit   *string  `json:"floorAreaUnit"`
	OccupancyStatus *string  `json:"occupancyStatus"`
}
