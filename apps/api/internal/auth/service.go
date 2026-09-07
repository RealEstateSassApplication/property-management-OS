package auth

import "context"

type MembershipRepository interface {
	GetMembership(ctx context.Context, organizationID, userID string) (Membership, error)
}

type Service struct {
	repository MembershipRepository
}

func NewService(repository MembershipRepository) *Service {
	return &Service{repository: repository}
}

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
	case "accountant", "viewer":
		return permission == ViewPortfolio || permission == ViewPeople || permission == ViewLeases
	case "maintenance":
		return permission == ViewPortfolio
	case "owner":
		// Owner-portal access must be resource-scoped to owned properties before
		// generalized organization endpoints are exposed to this role.
		return false
	default:
		return false
	}
}
