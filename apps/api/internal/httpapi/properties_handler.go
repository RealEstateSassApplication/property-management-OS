package httpapi

import (
	"errors"
	"net/http"

	"github.com/RealEstateSassApplication/property-management-OS/apps/api/internal/properties"
)

type propertyHandler struct {
	service *properties.Service
}

func (h propertyHandler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	items, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "properties_list_failed", "could not load properties")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": items})
}

func (h propertyHandler) get(w http.ResponseWriter, r *http.Request) {
	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Get(r.Context(), organizationID, r.PathValue("propertyID"))
	if errors.Is(err, properties.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property_not_found", "property not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "property_get_failed", "could not load property")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}

func (h propertyHandler) create(w http.ResponseWriter, r *http.Request) {
	var input properties.CreateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Create(r.Context(), organizationID, input)
	if err != nil {
		writeError(w, http.StatusBadRequest, "property_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": item})
}

func (h propertyHandler) update(w http.ResponseWriter, r *http.Request) {
	var input properties.UpdateInput
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}

	organizationID, _ := organizationIDFromContext(r.Context())
	item, err := h.service.Update(r.Context(), organizationID, r.PathValue("propertyID"), input)
	if errors.Is(err, properties.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property_not_found", "property not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "property_invalid", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": item})
}
