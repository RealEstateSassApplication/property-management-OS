package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenants"
)

type tenantHandler struct {
	service *tenants.Service
}

func (h tenantHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenants_list_failed", "could not load tenants")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h tenantHandler) get(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Get(r.Context(), organizationID, r.PathValue("tenantID"))
	if errors.Is(err, tenants.ErrNotFound) {
		writeError(w, http.StatusNotFound, "tenant_not_found", "tenant not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenant_get_failed", "could not load tenant")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h tenantHandler) create(w http.ResponseWriter, r *http.Request) {
	var input tenants.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Create(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tenant_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h tenantHandler) update(w http.ResponseWriter, r *http.Request) {
	var input tenants.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Update(r.Context(), organizationID, r.PathValue("tenantID"), input)
	if errors.Is(err, tenants.ErrNotFound) {
		writeError(w, http.StatusNotFound, "tenant_not_found", "tenant not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "tenant_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
