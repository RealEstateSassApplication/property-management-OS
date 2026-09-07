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
	case "admin", "manager":
		return true
	case "accountant":
		return permission == ViewPortfolio || permission == ViewPeople || permission == ViewLeases || permission == ViewOwners || permission == ViewRent || permission == ManageRent || permission == ViewMaintenance || permission == ApproveMaintenanceCosts || permission == ViewNotifications || permission == SendRentReminders || permission == ViewAgentActions || permission == DecideAgentActions
	case "viewer":
		return permission == ViewPortfolio || permission == ViewPeople || permission == ViewLeases || permission == ViewOwners || permission == ViewRent || permission == ViewMaintenance
	case "maintenance":
		return permission == ViewPortfolio || permission == ViewMaintenance || permission == CreateMaintenanceRequests || permission == ManageMaintenance || permission == ProposeAgentActions
	case "agent":
		return permission == ViewPortfolio || permission == ViewPeople || permission == ViewLeases || permission == ViewRent || permission == ViewMaintenance || permission == CreateMaintenanceRequests || permission == ViewNotifications || permission == SendRentReminders || permission == ViewAgentActions || permission == ProposeAgentActions
	case "owner":
		return false
	default:
		return false
	}
}
