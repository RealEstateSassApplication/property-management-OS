package orgadmin

import (
	"context"
	"fmt"
	"strings"
)

type Repository interface {
	GetSettings(ctx context.Context, organizationID string) (OrganizationSettings, error)
	UpdateSettings(ctx context.Context, organizationID string, input UpdateOrganizationInput) (OrganizationSettings, error)
	ListMembers(ctx context.Context, organizationID string) ([]Member, error)
	GetMember(ctx context.Context, organizationID, userID string) (Member, error)
	UpsertMember(ctx context.Context, organizationID string, input CreateMemberInput) (Member, error)
	UpdateMemberRole(ctx context.Context, organizationID, userID, role string) (Member, error)
	RemoveMember(ctx context.Context, organizationID, userID string) error
	CountAdmins(ctx context.Context, organizationID string) (int, error)
	ListPortalLinks(ctx context.Context, organizationID string) ([]PortalLink, error)
	LinkOwner(ctx context.Context, organizationID string, input CreateOwnerPortalLinkInput) (PortalLink, error)
	LinkTenant(ctx context.Context, organizationID string, input CreateTenantPortalLinkInput) (PortalLink, error)
	UnlinkOwner(ctx context.Context, organizationID, userID, ownerID string) error
	UnlinkTenant(ctx context.Context, organizationID, userID, tenantID string) error
}

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) GetSettings(ctx context.Context, organizationID string) (OrganizationSettings, error) {
	return s.repository.GetSettings(ctx, organizationID)
}

func (s *Service) UpdateSettings(ctx context.Context, organizationID string, input UpdateOrganizationInput) (OrganizationSettings, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Timezone = strings.TrimSpace(input.Timezone)
	input.DefaultCurrency = strings.ToUpper(strings.TrimSpace(input.DefaultCurrency))
	input.CountryCode = strings.ToUpper(strings.TrimSpace(input.CountryCode))
	input.BillingEmail = strings.ToLower(strings.TrimSpace(input.BillingEmail))
	if input.Name == "" || input.Timezone == "" {
		return OrganizationSettings{}, fmt.Errorf("name and timezone are required")
	}
	if len(input.DefaultCurrency) != 3 {
		return OrganizationSettings{}, fmt.Errorf("defaultCurrency must be a 3-letter code")
	}
	if len(input.CountryCode) != 2 {
		return OrganizationSettings{}, fmt.Errorf("countryCode must be a 2-letter code")
	}
	return s.repository.UpdateSettings(ctx, organizationID, input)
}

func (s *Service) ListMembers(ctx context.Context, organizationID string) ([]Member, error) {
	return s.repository.ListMembers(ctx, organizationID)
}

func (s *Service) CreateMember(ctx context.Context, organizationID string, input CreateMemberInput) (Member, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	input.Role = strings.ToLower(strings.TrimSpace(input.Role))
	if input.Email == "" || !strings.Contains(input.Email, "@") || input.DisplayName == "" {
		return Member{}, fmt.Errorf("valid email and displayName are required")
	}
	if !validRole(input.Role) {
		return Member{}, ErrInvalidRole
	}
	return s.repository.UpsertMember(ctx, organizationID, input)
}

func (s *Service) UpdateMemberRole(ctx context.Context, organizationID, userID string, input UpdateMemberRoleInput) (Member, error) {
	role := strings.ToLower(strings.TrimSpace(input.Role))
	if !validRole(role) {
		return Member{}, ErrInvalidRole
	}
	member, err := s.repository.GetMember(ctx, organizationID, strings.TrimSpace(userID))
	if err != nil {
		return Member{}, err
	}
	if member.Role == "admin" && role != "admin" {
		if err := s.ensureAnotherAdmin(ctx, organizationID); err != nil {
			return Member{}, err
		}
	}
	return s.repository.UpdateMemberRole(ctx, organizationID, userID, role)
}

func (s *Service) RemoveMember(ctx context.Context, organizationID, userID string) error {
	member, err := s.repository.GetMember(ctx, organizationID, strings.TrimSpace(userID))
	if err != nil {
		return err
	}
	if member.Role == "admin" {
		if err := s.ensureAnotherAdmin(ctx, organizationID); err != nil {
			return err
		}
	}
	return s.repository.RemoveMember(ctx, organizationID, userID)
}

func (s *Service) ListPortalLinks(ctx context.Context, organizationID string) ([]PortalLink, error) {
	return s.repository.ListPortalLinks(ctx, organizationID)
}

func (s *Service) LinkOwner(ctx context.Context, organizationID string, input CreateOwnerPortalLinkInput) (PortalLink, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.OwnerID = strings.TrimSpace(input.OwnerID)
	if input.UserID == "" || input.OwnerID == "" {
		return PortalLink{}, fmt.Errorf("userId and ownerId are required")
	}
	return s.repository.LinkOwner(ctx, organizationID, input)
}

func (s *Service) LinkTenant(ctx context.Context, organizationID string, input CreateTenantPortalLinkInput) (PortalLink, error) {
	input.UserID = strings.TrimSpace(input.UserID)
	input.TenantID = strings.TrimSpace(input.TenantID)
	if input.UserID == "" || input.TenantID == "" {
		return PortalLink{}, fmt.Errorf("userId and tenantId are required")
	}
	return s.repository.LinkTenant(ctx, organizationID, input)
}

func (s *Service) UnlinkOwner(ctx context.Context, organizationID, userID, ownerID string) error {
	return s.repository.UnlinkOwner(ctx, organizationID, strings.TrimSpace(userID), strings.TrimSpace(ownerID))
}

func (s *Service) UnlinkTenant(ctx context.Context, organizationID, userID, tenantID string) error {
	return s.repository.UnlinkTenant(ctx, organizationID, strings.TrimSpace(userID), strings.TrimSpace(tenantID))
}

func (s *Service) ensureAnotherAdmin(ctx context.Context, organizationID string) error {
	count, err := s.repository.CountAdmins(ctx, organizationID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return ErrLastAdmin
	}
	return nil
}

func validRole(role string) bool {
	switch role {
	case "admin", "manager", "accountant", "maintenance", "viewer", "owner", "tenant", "agent":
		return true
	default:
		return false
	}
}
