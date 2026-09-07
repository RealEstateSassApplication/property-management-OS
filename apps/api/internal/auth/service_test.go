package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeMembershipRepository struct {
	membership Membership
	err        error
}

func (f fakeMembershipRepository) GetMembership(context.Context, string, string) (Membership, error) {
	return f.membership, f.err
}

func TestAuthorizeAllowsManager(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "manager"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ManageLeases); err != nil {
		t.Fatalf("expected manager to be allowed: %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ManageNotifications); err != nil {
		t.Fatalf("expected manager to manage notifications: %v", err)
	}
}

func TestAuthorizeAllowsAccountantToManageRentAndApproveMaintenanceCosts(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "accountant"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ManageRent); err != nil {
		t.Fatalf("expected accountant to manage rent: %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ApproveMaintenanceCosts); err != nil {
		t.Fatalf("expected accountant to approve maintenance costs: %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ManageMaintenance); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected accountant operational maintenance writes to be forbidden, got %v", err)
	}
}

func TestAuthorizeAccountantCanQueueRentReminderButNotArbitraryNotification(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "accountant"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ViewNotifications); err != nil {
		t.Fatalf("expected accountant to view notification delivery state: %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", SendRentReminders); err != nil {
		t.Fatalf("expected accountant to send server-authored rent reminders: %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ManageNotifications); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected generic notification creation to remain manager/admin-only, got %v", err)
	}
}

func TestAuthorizeAllowsMaintenanceOperationsButNotCostOrVendorApproval(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "maintenance"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ManageMaintenance); err != nil {
		t.Fatalf("expected maintenance role to operate work orders: %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ApproveMaintenanceCosts); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected maintenance cost approval to be forbidden, got %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ManageMaintenanceVendors); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected maintenance vendor management to be forbidden, got %v", err)
	}
}

func TestAuthorizeRestrictsViewerWrites(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "viewer"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ManageRent); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ViewMaintenance); err != nil {
		t.Fatalf("expected viewer to read maintenance, got %v", err)
	}
	if _, err := service.Authorize(context.Background(), "org", "user", ViewNotifications); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected notification delivery register to stay restricted, got %v", err)
	}
}

func TestAuthorizeDoesNotGrantGeneralizedOwnerAccess(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "owner"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ViewPortfolio); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected owner role to require resource-scoped portal authorization, got %v", err)
	}
}
