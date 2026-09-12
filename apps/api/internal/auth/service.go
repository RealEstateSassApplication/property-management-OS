package auth

import "context"

type MembershipRepository interface {
	GetMembership(ctx context.Context, organizationID, userID string) (Membership, error)
}

type Service struct{ repository MembershipRepository }

func NewService(repository MembershipRepository) *Service { return &Service{repository: repository} }

func (s *Service) Authorize(ctx context.Context, organizationID, userID string, permission Permission) (Membership, error) {
	membership, err := s.repository.GetMembership(ctx, organizationID, userID)
	if err != nil {
		return Membership{}, err
	}
	if !roleAllows(membership.Role, permission) {
		return Membership{}, ErrForbidden
	}
	return membership, nil
}

func roleAllows(role string, permission Permission) bool {
	switch role {
	case "admin":
		return true
	case "manager":
		return permission != ManageMembers && permission != ManageAccounting
	case "accountant":
		return permission == ViewPortfolio || permission == ViewPeople || permission == ViewLeases || permission == ViewOwners || permission == ViewRent || permission == ManageRent || permission == ViewAccounting || permission == ManageAccounting || permission == ViewInspections || permission == ViewMaintenance || permission == ApproveMaintenanceCosts || permission == ViewNotifications || permission == ManageOwnPushDevices || permission == SendRentReminders || permission == ViewAgentActions || permission == DecideAgentActions || permission == ViewReporting || permission == ViewAudit
	case "viewer":
		return permission == ViewPortfolio || permission == ViewPeople || permission == ViewLeases || permission == ViewOwners || permission == ViewRent || permission == ViewInspections || permission == ViewMaintenance || permission == ViewReporting || permission == ManageOwnPushDevices
	case "maintenance":
		return permission == ViewPortfolio || permission == ViewInspections || permission == ManageInspections || permission == AcknowledgeInspections || permission == ViewMaintenance || permission == CreateMaintenanceRequests || permission == ManageMaintenance || permission == ProposeAgentActions || permission == ManageOwnPushDevices
	case "agent":
		return permission == ViewPortfolio || permission == ViewPeople || permission == ViewLeases || permission == ViewRent || permission == ViewInspections || permission == ViewMaintenance || permission == CreateMaintenanceRequests || permission == ViewNotifications || permission == SendRentReminders || permission == ViewAgentActions || permission == ProposeAgentActions || permission == ManageOwnPushDevices
	case "owner":
		return permission == ViewOwnerPortal || permission == ManageOwnPushDevices
	case "tenant":
		return permission == ViewTenantPortal || permission == CreateTenantPortalRequest || permission == ManageOwnPushDevices
	default:
		return false
	}
}
