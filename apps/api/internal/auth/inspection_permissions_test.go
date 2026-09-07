package auth

import (
	"context"
	"errors"
	"testing"
)

func TestInspectionPermissions(t *testing.T) {
	manager := NewService(fakeMembershipRepository{membership: Membership{Role: "manager"}})
	for _, permission := range []Permission{ViewInspections, ManageInspections, AcknowledgeInspections} {
		if _, err := manager.Authorize(context.Background(), "org", "manager", permission); err != nil {
			t.Fatalf("manager should have %s: %v", permission, err)
		}
	}

	viewer := NewService(fakeMembershipRepository{membership: Membership{Role: "viewer"}})
	if _, err := viewer.Authorize(context.Background(), "org", "viewer", ViewInspections); err != nil {
		t.Fatalf("viewer should read inspections: %v", err)
	}
	if _, err := viewer.Authorize(context.Background(), "org", "viewer", ManageInspections); !errors.Is(err, ErrForbidden) {
		t.Fatalf("viewer should not manage inspections, got %v", err)
	}

	tenant := NewService(fakeMembershipRepository{membership: Membership{Role: "tenant"}})
	if _, err := tenant.Authorize(context.Background(), "org", "tenant", ViewInspections); !errors.Is(err, ErrForbidden) {
		t.Fatalf("tenant must not get organization-wide inspection access, got %v", err)
	}
}
