package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/units"
)

type unitHandler struct {
	service *units.Service
}

func (h unitHandler) listByProperty(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.ListByProperty(r.Context(), organizationID, r.PathValue("propertyID"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "units_list_failed", "could not load units")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h unitHandler) get(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Get(r.Context(), organizationID, r.PathValue("unitID"))
	if errors.Is(err, units.ErrNotFound) {
		writeError(w, http.StatusNotFound, "unit_not_found", "unit not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "unit_get_failed", "could not load unit")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h unitHandler) create(w http.ResponseWriter, r *http.Request) {
	var input units.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Create(r.Context(), organizationID, r.PathValue("propertyID"), input)
	if errors.Is(err, units.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property_not_found", "property not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "unit_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h unitHandler) update(w http.ResponseWriter, r *http.Request) {
	var input units.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Update(r.Context(), organizationID, r.PathValue("unitID"), input)
	if errors.Is(err, units.ErrNotFound) {
		writeError(w, http.StatusNotFound, "unit_not_found", "unit not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "unit_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
