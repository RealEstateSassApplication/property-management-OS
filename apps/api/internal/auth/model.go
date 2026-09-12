package auth

import "errors"

type Permission string

const (
	ViewPortfolio             Permission = "portfolio:view"
	ManagePortfolio           Permission = "portfolio:manage"
	ViewPeople                Permission = "people:view"
	ManagePeople              Permission = "people:manage"
	ViewLeases                Permission = "leases:view"
	ManageLeases              Permission = "leases:manage"
	ViewOwners                Permission = "owners:view"
	ManageOwners              Permission = "owners:manage"
	ViewRent                  Permission = "rent:view"
	ManageRent                Permission = "rent:manage"
	ViewAccounting            Permission = "accounting:view"
	ManageAccounting          Permission = "accounting:manage"
	ViewInspections           Permission = "inspections:view"
	ManageInspections         Permission = "inspections:manage"
	AcknowledgeInspections    Permission = "inspections:acknowledge"
	ViewMaintenance           Permission = "maintenance:view"
	CreateMaintenanceRequests Permission = "maintenance:create_requests"
	ManageMaintenance         Permission = "maintenance:manage"
	ManageMaintenanceVendors  Permission = "maintenance:manage_vendors"
	ApproveMaintenanceCosts   Permission = "maintenance:approve_costs"
	ViewDocuments             Permission = "documents:view"
	ManageDocuments           Permission = "documents:manage"
	ViewNotifications         Permission = "notifications:view"
	ManageNotifications       Permission = "notifications:manage"
	ManageOwnPushDevices      Permission = "notifications:push_devices:self"
	SendRentReminders         Permission = "notifications:send_rent_reminders"
	ViewAgentActions          Permission = "agent_actions:view"
	ProposeAgentActions       Permission = "agent_actions:propose"
	DecideAgentActions        Permission = "agent_actions:decide"
	ViewOwnerPortal           Permission = "portal:owner:view"
	ViewTenantPortal          Permission = "portal:tenant:view"
	CreateTenantPortalRequest Permission = "portal:tenant:create_maintenance_request"
	ViewOrganization          Permission = "organization:view"
	ManageOrganization        Permission = "organization:manage"
	ViewMembers               Permission = "organization:members:view"
	ManageMembers             Permission = "organization:members:manage"
	ManagePortalLinks         Permission = "organization:portal_links:manage"
	ViewReporting             Permission = "reporting:view"
	ViewAudit                 Permission = "audit:view"
)

type Membership struct {
	OrganizationID string `json:"organizationId"`
	UserID         string `json:"userId"`
	Role           string `json:"role"`
}

var (
	ErrMembershipNotFound = errors.New("organization membership not found")
	ErrForbidden          = errors.New("permission denied")
)
