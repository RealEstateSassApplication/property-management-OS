package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/tenancies"
)

type tenancyHandler struct {
	service *tenancies.Service
}

func (h tenancyHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenancies_list_failed", "could not load tenancies")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h tenancyHandler) get(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Get(r.Context(), organizationID, r.PathValue("tenancyID"))
	if errors.Is(err, tenancies.ErrNotFound) {
		writeError(w, http.StatusNotFound, "tenancy_not_found", "tenancy not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "tenancy_get_failed", "could not load tenancy")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h tenancyHandler) create(w http.ResponseWriter, r *http.Request) {
	var input tenancies.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Create(r.Context(), organizationID, input)
	switch {
	case errors.Is(err, tenancies.ErrUnitNotFound), errors.Is(err, tenancies.ErrTenantNotFound):
		writeError(w, http.StatusNotFound, "tenancy_dependency_not_found", "unit or tenant not found")
		return
	case errors.Is(err, tenancies.ErrUnitConflict):
		writeError(w, http.StatusConflict, "unit_already_occupied", "unit already has an active tenancy")
		return
	case errors.Is(err, tenancies.ErrTenantInvalid):
		writeError(w, http.StatusUnprocessableEntity, "tenant_not_eligible", "tenant is not eligible for tenancy")
		return
	case err != nil:
		writeError(w, http.StatusBadRequest, "tenancy_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h tenancyHandler) update(w http.ResponseWriter, r *http.Request) {
	var input tenancies.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Update(r.Context(), organizationID, r.PathValue("tenancyID"), input)
	if errors.Is(err, tenancies.ErrNotFound) {
		writeError(w, http.StatusNotFound, "tenancy_not_found", "tenancy not found")
		return
	}
	if errors.Is(err, tenancies.ErrUnitConflict) {
		writeError(w, http.StatusConflict, "unit_already_occupied", "unit already has an active tenancy")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "tenancy_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
