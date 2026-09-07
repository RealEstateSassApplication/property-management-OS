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
}

func TestAuthorizeRestrictsViewerWrites(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "viewer"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ManagePeople); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestAuthorizeDoesNotGrantGeneralizedOwnerAccess(t *testing.T) {
	service := NewService(fakeMembershipRepository{membership: Membership{Role: "owner"}})
	if _, err := service.Authorize(context.Background(), "org", "user", ViewPortfolio); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected owner role to require resource-scoped portal authorization, got %v", err)
	}
}
