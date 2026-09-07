package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/portals"
)

type portalHandler struct{ service *portals.Service }

func (h portalHandler) ownerSummary(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.OwnerSummary(r.Context(), organizationID, userID)
	switch {
	case errors.Is(err, portals.ErrOwnerLinkNotFound):
		writeError(w, http.StatusNotFound, "owner_portal_not_linked", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "owner_portal_failed", "could not load owner portal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h portalHandler) tenantSummary(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.TenantSummary(r.Context(), organizationID, userID)
	switch {
	case errors.Is(err, portals.ErrTenantLinkNotFound):
		writeError(w, http.StatusNotFound, "tenant_portal_not_linked", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "tenant_portal_failed", "could not load tenant portal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h portalHandler) createTenantMaintenance(w http.ResponseWriter, r *http.Request) {
	var input portals.CreateTenantMaintenanceInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	userID, _ := userIDFromContext(r.Context())
	item, err := h.service.CreateTenantMaintenance(r.Context(), organizationID, userID, input)
	switch {
	case errors.Is(err, portals.ErrTenantLinkNotFound):
		writeError(w, http.StatusNotFound, "tenant_portal_not_linked", err.Error())
		return
	case errors.Is(err, portals.ErrTenancyNotAccessible):
		writeError(w, http.StatusForbidden, "tenancy_not_accessible", err.Error())
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "tenant_maintenance_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}
