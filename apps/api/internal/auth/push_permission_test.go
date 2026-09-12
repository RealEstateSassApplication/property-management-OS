package auth

import "testing"

func TestEveryOrganizationRoleCanManageOwnPushDevices(t *testing.T) {
	roles := []string{"admin", "manager", "accountant", "viewer", "maintenance", "agent", "owner", "tenant"}
	for _, role := range roles {
		if !roleAllows(role, ManageOwnPushDevices) {
			t.Fatalf("role %q should be allowed to manage its own push devices", role)
		}
	}
}

func TestPushDevicePermissionDoesNotGrantNotificationManagement(t *testing.T) {
	for _, role := range []string{"viewer", "maintenance", "agent", "owner", "tenant"} {
		if roleAllows(role, ManageNotifications) {
			t.Fatalf("role %q unexpectedly has notification management permission", role)
		}
	}
}
