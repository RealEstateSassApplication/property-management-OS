package reporting

import (
	"encoding/json"
	"time"
)

type CurrencyAmount struct {
	Currency    string `json:"currency"`
	AmountMinor int64  `json:"amountMinor"`
}

type Dashboard struct {
	PropertyCount            int              `json:"propertyCount"`
	UnitCount                int              `json:"unitCount"`
	OccupiedUnits            int              `json:"occupiedUnits"`
	VacancyRateBPS           int              `json:"vacancyRateBps"`
	OpenMaintenance          int              `json:"openMaintenance"`
	EmergencyMaintenance     int              `json:"emergencyMaintenance"`
	LeasesExpiring30Days     int              `json:"leasesExpiring30Days"`
	LeasesExpiring90Days     int              `json:"leasesExpiring90Days"`
	OutstandingByCurrency    []CurrencyAmount `json:"outstandingByCurrency"`
	OverdueByCurrency        []CurrencyAmount `json:"overdueByCurrency"`
	CollectedMonthByCurrency []CurrencyAmount `json:"collectedThisMonthByCurrency"`
}

type AuditEvent struct {
	ID             string          `json:"id"`
	ActorUserID    string          `json:"actorUserId,omitempty"`
	ActorName      string          `json:"actorName,omitempty"`
	Action         string          `json:"action"`
	ResourceType   string          `json:"resourceType"`
	ResourceID     string          `json:"resourceId,omitempty"`
	RequestID      string          `json:"requestId,omitempty"`
	Metadata       json.RawMessage `json:"metadata"`
	OccurredAt     time.Time       `json:"occurredAt"`
}
