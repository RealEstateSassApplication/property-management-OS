package orgadmin

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	settings    OrganizationSettings
	members     []Member
	member      Member
	adminCount  int
	updatedRole string
	removedUser string
}

func (f *fakeRepository) GetSettings(context.Context, string) (OrganizationSettings, error) { return f.settings, nil }
func (f *fakeRepository) UpdateSettings(_ context.Context, _ string, input UpdateOrganizationInput) (OrganizationSettings, error) {
	f.settings.Name = input.Name
	f.settings.Timezone = input.Timezone
	f.settings.DefaultCurrency = input.DefaultCurrency
	f.settings.CountryCode = input.CountryCode
	f.settings.BillingEmail = input.BillingEmail
	return f.settings, nil
}
func (f *fakeRepository) ListMembers(context.Context, string) ([]Member, error) { return f.members, nil }
func (f *fakeRepository) GetMember(context.Context, string, string) (Member, error) { return f.member, nil }
func (f *fakeRepository) UpsertMember(_ context.Context, _ string, input CreateMemberInput) (Member, error) {
	return Member{Email: input.Email, DisplayName: input.DisplayName, Role: input.Role}, nil
}
func (f *fakeRepository) UpdateMemberRole(_ context.Context, _, userID, role string) (Member, error) {
	f.updatedRole = role
	return Member{UserID: userID, Role: role}, nil
}
func (f *fakeRepository) RemoveMember(_ context.Context, _, userID string) error { f.removedUser = userID; return nil }
func (f *fakeRepository) CountAdmins(context.Context, string) (int, error) { return f.adminCount, nil }
func (f *fakeRepository) ListPortalLinks(context.Context, string) ([]PortalLink, error) { return nil, nil }
func (f *fakeRepository) LinkOwner(_ context.Context, _ string, input CreateOwnerPortalLinkInput) (PortalLink, error) {
	return PortalLink{Kind: "owner", UserID: input.UserID, ResourceID: input.OwnerID}, nil
}
func (f *fakeRepository) LinkTenant(_ context.Context, _ string, input CreateTenantPortalLinkInput) (PortalLink, error) {
	return PortalLink{Kind: "tenant", UserID: input.UserID, ResourceID: input.TenantID}, nil
}
func (f *fakeRepository) UnlinkOwner(context.Context, string, string, string) error { return nil }
func (f *fakeRepository) UnlinkTenant(context.Context, string, string, string) error { return nil }

func TestUpdateSettingsNormalizesCodes(t *testing.T) {
	repo := &fakeRepository{}
	service := NewService(repo)
	item, err := service.UpdateSettings(context.Background(), "org", UpdateOrganizationInput{Name: " Avara Ops ", Timezone: " Asia/Colombo ", DefaultCurrency: "lkr", CountryCode: "lk", BillingEmail: " BILLING@EXAMPLE.COM "})
	if err != nil { t.Fatal(err) }
	if item.Name != "Avara Ops" || item.DefaultCurrency != "LKR" || item.CountryCode != "LK" || item.BillingEmail != "billing@example.com" {
		t.Fatalf("unexpected normalized settings: %+v", item)
	}
}

func TestFinalAdminCannotBeDemoted(t *testing.T) {
	repo := &fakeRepository{member: Member{UserID: "admin-1", Role: "admin"}, adminCount: 1}
	service := NewService(repo)
	_, err := service.UpdateMemberRole(context.Background(), "org", "admin-1", UpdateMemberRoleInput{Role: "manager"})
	if !errors.Is(err, ErrLastAdmin) { t.Fatalf("expected ErrLastAdmin, got %v", err) }
	if repo.updatedRole != "" { t.Fatalf("final admin role was mutated") }
}

func TestAdminCanBeDemotedWhenAnotherAdminExists(t *testing.T) {
	repo := &fakeRepository{member: Member{UserID: "admin-1", Role: "admin"}, adminCount: 2}
	service := NewService(repo)
	item, err := service.UpdateMemberRole(context.Background(), "org", "admin-1", UpdateMemberRoleInput{Role: "manager"})
	if err != nil { t.Fatal(err) }
	if item.Role != "manager" || repo.updatedRole != "manager" { t.Fatalf("expected manager role, got %+v", item) }
}

func TestFinalAdminCannotBeRemoved(t *testing.T) {
	repo := &fakeRepository{member: Member{UserID: "admin-1", Role: "admin"}, adminCount: 1}
	service := NewService(repo)
	err := service.RemoveMember(context.Background(), "org", "admin-1")
	if !errors.Is(err, ErrLastAdmin) { t.Fatalf("expected ErrLastAdmin, got %v", err) }
	if repo.removedUser != "" { t.Fatalf("final admin was removed") }
}

func TestCreateMemberRejectsUnsupportedRole(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.CreateMember(context.Background(), "org", CreateMemberInput{Email: "user@example.com", DisplayName: "User", Role: "superadmin"})
	if !errors.Is(err, ErrInvalidRole) { t.Fatalf("expected ErrInvalidRole, got %v", err) }
}
