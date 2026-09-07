package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/leases"
)

type leaseHandler struct {
	service *leases.Service
}

func (h leaseHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "leases_list_failed", "could not load leases")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h leaseHandler) get(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Get(r.Context(), organizationID, r.PathValue("leaseID"))
	if errors.Is(err, leases.ErrNotFound) {
		writeError(w, http.StatusNotFound, "lease_not_found", "lease not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "lease_get_failed", "could not load lease")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h leaseHandler) create(w http.ResponseWriter, r *http.Request) {
	var input leases.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Create(r.Context(), organizationID, input)
	if errors.Is(err, leases.ErrTenancyNotFound) {
		writeError(w, http.StatusNotFound, "tenancy_not_found", "tenancy not found")
		return
	}
	if errors.Is(err, leases.ErrActiveLeaseConflict) {
		writeError(w, http.StatusConflict, "active_lease_exists", "tenancy already has an active lease")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "lease_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h leaseHandler) update(w http.ResponseWriter, r *http.Request) {
	var input leases.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Update(r.Context(), organizationID, r.PathValue("leaseID"), input)
	if errors.Is(err, leases.ErrNotFound) {
		writeError(w, http.StatusNotFound, "lease_not_found", "lease not found")
		return
	}
	if errors.Is(err, leases.ErrActiveLeaseConflict) {
		writeError(w, http.StatusConflict, "active_lease_exists", "tenancy already has an active lease")
		return
	}
	if errors.Is(err, leases.ErrInvalidTransition) {
		writeError(w, http.StatusConflict, "invalid_lease_transition", "lease status transition is not allowed")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "lease_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
